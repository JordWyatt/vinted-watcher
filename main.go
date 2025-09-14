package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "vinted-watcher/internal/logger"
	"vinted-watcher/internal/scraper"
	"vinted-watcher/internal/server"
	"vinted-watcher/internal/storage"
	"vinted-watcher/internal/vinted"
)

const DISCORD_WEBHOOK_URL_ENV_VAR = "DISCORD_WEBHOOK_URL"
const DB_PATH_ENV_VAR = "DB_PATH"
const DEFAULT_DB_PATH = "./vinted.db"
const VINTED_BASE_URL = "http://www.vinted.co.uk"

// Test code - will eventually become server entrypoint
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		slog.Info("Received shutdown signal, shutting down gracefully...")
		cancel()
	}()

	db, err := storage.NewDB(getEnvVar(DB_PATH_ENV_VAR, DEFAULT_DB_PATH))
	if err != nil {
		slog.Error("Error initializing database", "error", err)
		return
	}

	vintedClient := vinted.NewClient(VINTED_BASE_URL)

	vintedScraper := scraper.NewScraper(vintedClient, db, scraper.ScraperConfig{
		LookbackPeriod:                24 * time.Hour,
		DiscordNotificationWebhookURL: os.Getenv(DISCORD_WEBHOOK_URL_ENV_VAR),
	})

	go startScheduler(ctx, vintedScraper, 1*time.Hour)

	httpServer := server.NewServer(db, vintedScraper)
	if err := httpServer.Start(ctx); err != nil {
		slog.Error("Error starting server:", "error", err)
	}
}

func startScheduler(ctx context.Context, scraper *scraper.Scraper, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	safeScrape(scraper) // Initial scrape on startup
	for {
		select {
		case <-ticker.C:
			safeScrape(scraper)
		case <-ctx.Done():
			slog.Info("Stopping scheduled scrape...")
			return
		}
	}
}

func safeScrape(scraper *scraper.Scraper) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Panic occurred during scraping:", "error", r)
		}
	}()
	scraperResult, err := scraper.Scrape()
	if err != nil {
		slog.Error("Error scraping:", "error", err)
		return
	}
	slog.Debug("Scrape stats", "new_item_count", len(scraperResult.NewItems), "processed_searches_count", scraperResult.ProcessedSearches)
}

func getEnvVar(varName string, defaultValue string) string {
	value := os.Getenv(varName)
	if value == "" {
		return defaultValue
	}
	return value
}
