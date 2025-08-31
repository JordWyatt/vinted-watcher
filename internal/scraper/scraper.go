package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"vinted-watcher/internal/discord"
	"vinted-watcher/internal/domain"
	"vinted-watcher/internal/storage"
	"vinted-watcher/internal/vinted"
)

const (
	maxEmbedsPerMessage = 10 // Discord limit
	notificationTimeout = 10 * time.Second
)

type ScraperConfig struct {
	LookbackPeriod                time.Duration
	DiscordNotificationWebhookURL string
}

type Scraper struct {
	vintedClient vinted.VintedClient
	db           storage.SearchStorage
	config       ScraperConfig
	discord      *discord.DiscordWebhook
}

type ScraperResult struct {
	NewItems          []vinted.Item
	ProcessedSearches int
	Errors            []error
}

func NewScraper(vintedClient vinted.VintedClient, db storage.SearchStorage, config ScraperConfig) *Scraper {
	s := &Scraper{
		vintedClient: vintedClient,
		db:           db,
		config:       config,
	}

	if config.DiscordNotificationWebhookURL != "" {
		s.discord = discord.NewDiscordWebhook(config.DiscordNotificationWebhookURL)
	}

	return s
}

func (s *Scraper) Scrape() (*ScraperResult, error) {

	slog.Info("Scraping...")
	result := &ScraperResult{
		NewItems: make([]vinted.Item, 0),
		Errors:   make([]error, 0),
	}

	activeSearches, err := s.getActiveSearches()
	if err != nil {
		return nil, fmt.Errorf("failed to get searches: %w", err)
	}

	for _, search := range activeSearches {
		newItems, err := s.processSearch(search)
		if err != nil {
			slog.Error("Error processing search", "search_id", search.ID, "err", err.Error())
			result.Errors = append(result.Errors, fmt.Errorf("search %d: %w", search.ID, err))
			continue
		}

		result.NewItems = append(result.NewItems, newItems...)
		result.ProcessedSearches++

		slog.Debug("Completed search", "search_id", search.ID, "new_items_count", len(newItems))
	}

	slog.Info("Scraping complete")
	return result, nil
}

func (s *Scraper) getActiveSearches() ([]domain.SavedSearch, error) {
	searches, err := s.db.GetAllSearches()
	if err != nil {
		return nil, fmt.Errorf("failed to get searches: %w", err)
	}

	activeSearches := make([]domain.SavedSearch, 0, len(searches))
	for _, search := range searches {
		if search.IsActive() {
			activeSearches = append(activeSearches, *search)
		}
	}

	return activeSearches, nil
}

func (s *Scraper) processSearch(search domain.SavedSearch) ([]vinted.Item, error) {
	items, err := s.getItemsForSearch(search)

	slog.Info("Items found", "count", len(items))
	if err != nil {
		return nil, err
	}

	recentItems := s.filterItemsByLookback(items)

	slog.Info("Items remaining after lookback filter", "count", len(recentItems))

	newItems := make([]vinted.Item, 0)

	for _, item := range recentItems {
		isNew, err := s.processItem(search, item)
		if err != nil {
			return nil, fmt.Errorf("failed to process item %d: %w", item.ID, err)
		}

		if isNew {
			newItems = append(newItems, item)
		}

		slog.Debug("Processed item", "item_id", item.ID, "search_id", search.ID)
	}

	if s.discord != nil && len(newItems) > 0 {
		slog.Info("posting discord notification for search", "search_id", search.ID)
		err := s.postDiscordNotification(newItems, search)
		if err != nil {
			return nil, fmt.Errorf("failed to post discord notification: %w", err)
		}
	}

	return newItems, nil
}

func (s *Scraper) getItemsForSearch(search domain.SavedSearch) ([]vinted.Item, error) {
	items, err := s.vintedClient.GetItems(search.SearchParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get items for search %d: %w", search.ID, err)
	}
	return items, nil
}

func (s *Scraper) processItem(search domain.SavedSearch, item vinted.Item) (bool, error) {
	seen, err := s.db.IsItemSeen(search.ID, int(item.ID))
	if err != nil {
		return false, fmt.Errorf("failed to check if item is seen: %w", err)
	}

	if seen {
		return false, nil
	}

	err = s.db.MarkItemAsSeen(search.ID, int(item.ID))
	if err != nil {
		return false, fmt.Errorf("failed to mark item as seen: %w", err)
	}

	return true, nil
}

// filterItemsByLookback filters items based on the configured lookback period
func (s *Scraper) filterItemsByLookback(items []vinted.Item) []vinted.Item {
	if len(items) == 0 {
		return items
	}

	cutoff := time.Now().Add(-s.config.LookbackPeriod)
	filteredItems := make([]vinted.Item, 0, len(items))

	for _, item := range items {
		if s.isItemWithinLookback(item, cutoff) {
			filteredItems = append(filteredItems, item)
		}
	}

	return filteredItems
}

// isItemWithinLookback checks if an item is within the lookback period
func (s *Scraper) isItemWithinLookback(item vinted.Item, cutoff time.Time) bool {
	uploadedAt := time.Unix(int64(item.Photo.HighResolution.Timestamp), 0)
	slog.Debug("Checking whether item was uploaded within lookback period", "item_name", item.Title, "item_id", item.ID, "uploaded_at", uploadedAt, "cutoff_time", cutoff)
	return uploadedAt.After(cutoff)
}

func (s *Scraper) postDiscordNotification(items []vinted.Item, search domain.SavedSearch) error {
	if len(items) == 0 {
		return nil // No items to notify about
	}

	// Split items into batches if needed (Discord has a limit of 10 embeds per message)
	batches := s.createItemBatches(items, maxEmbedsPerMessage)

	ctx, cancel := context.WithTimeout(context.Background(), notificationTimeout)
	defer cancel()

	for i, batch := range batches {
		if err := s.sendBatch(ctx, batch, search, i, len(batches)); err != nil {
			return fmt.Errorf("failed to send batch %d: %w", i+1, err)
		}
	}

	return nil
}

func (s *Scraper) createItemBatches(items []vinted.Item, batchSize int) [][]vinted.Item {
	var batches [][]vinted.Item

	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batches = append(batches, items[i:end])
	}

	return batches
}

func (s *Scraper) sendBatch(ctx context.Context, items []vinted.Item, search domain.SavedSearch, batchNum, totalBatches int) error {
	message := s.createDiscordMessage(items, search, batchNum, totalBatches)

	if err := s.discord.PostMessage(ctx, message); err != nil {
		return fmt.Errorf("discord API error: %w", err)
	}

	return nil
}

func (s *Scraper) createDiscordMessage(items []vinted.Item, search domain.SavedSearch, batchNum, totalBatches int) discord.WebhookMessage {
	content := s.formatMessageContent(search, len(items), batchNum, totalBatches)

	message := discord.WebhookMessage{
		Content: content,
		Embeds:  make([]discord.Embed, 0, len(items)),
	}

	for _, item := range items {
		embed := s.createItemEmbed(item)
		message.Embeds = append(message.Embeds, embed)
	}

	return message
}

func (s *Scraper) formatMessageContent(search domain.SavedSearch, itemCount, batchNum, totalBatches int) string {
	if totalBatches == 1 {
		return fmt.Sprintf("🔍 **%s**: %d new item(s) found", search.Name, itemCount)
	}

	return fmt.Sprintf("🔍 **%s**: Batch %d/%d", search.Name, batchNum+1, totalBatches)
}

func (s *Scraper) createItemEmbed(item vinted.Item) discord.Embed {
	embed := discord.Embed{
		Title: s.truncateTitle(item.Title, 256), // Discord title limit
		URL:   item.URL,
		Fields: []discord.EmbedField{
			{
				Name:   "💰 Price",
				Value:  s.formatPrice(item.Price),
				Inline: true,
			},
		},
	}

	// Add size field if available
	if item.SizeTitle != "" {
		embed.Fields = append(embed.Fields, discord.EmbedField{
			Name:   "📏 Size",
			Value:  item.SizeTitle,
			Inline: true,
		})
	}

	// Add brand field if available
	if item.BrandTitle != "" {
		embed.Fields = append(embed.Fields, discord.EmbedField{
			Name:   "🏷️ Brand",
			Value:  item.BrandTitle,
			Inline: true,
		})
	}

	// Add image if available
	if item.Photo.URL != "" {
		embed.Image = discord.EmbedImage{
			URL: item.Photo.URL,
		}
	}

	return embed
}

func (s *Scraper) formatPrice(price vinted.Price) string {
	if price.Amount == "" || price.CurrencyCode == "" {
		return "Price not available"
	}

	// Handle different currency formats
	switch price.CurrencyCode {
	case "EUR":
		return fmt.Sprintf("€%s", price.Amount)
	case "USD":
		return fmt.Sprintf("$%s", price.Amount)
	case "GBP":
		return fmt.Sprintf("£%s", price.Amount)
	default:
		return fmt.Sprintf("%s %s", price.Amount, price.CurrencyCode)
	}
}

func (s *Scraper) truncateTitle(title string, maxLength int) string {
	if len(title) <= maxLength {
		return title
	}
	return title[:maxLength-3] + "..."
}
