package scraper

import (
	"fmt"
	"log/slog"
	"time"
	"vinted-watcher/internal/domain"
	"vinted-watcher/internal/storage"
	"vinted-watcher/internal/vinted"
)

type ScraperConfig struct {
	LookbackPeriod time.Duration
}

type Scraper struct {
	vintedClient vinted.VintedClient
	db           storage.SearchStorage
	config       ScraperConfig
}

type ScraperResult struct {
	NewItems          []vinted.Item
	ProcessedSearches int
	Errors            []error
}

func NewScraper(vintedClient vinted.VintedClient, db storage.SearchStorage, config ScraperConfig) *Scraper {
	return &Scraper{
		vintedClient: vintedClient,
		db:           db,
		config:       config,
	}
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
	slog.Debug("Checking whether item was uploaded within lookback period", "item_id", item.ID, "uploaded_at", item.Photo.HighResolution.Timestamp, "cutoff_time", cutoff)
	uploadedAt := time.Unix(int64(item.Photo.HighResolution.Timestamp), 0)
	return uploadedAt.After(cutoff)
}
