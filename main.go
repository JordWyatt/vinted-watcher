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

	db, err := storage.NewDB("vinted.db")
	if err != nil {
		slog.Error("Error initializing database", "error", err)
		return
	}

	vintedClient := vinted.NewClient("https://www.vinted.com")

	scraper := scraper.NewScraper(vintedClient, db, scraper.ScraperConfig{
		LookbackPeriod: 24 * time.Hour,
	})

	go startScheduler(ctx, scraper, 1*time.Minute)

	server := server.NewServer(db)
	if err := server.Start(ctx); err != nil {
		slog.Error("Error starting server:", "error", err)
	}
}

func startScheduler(ctx context.Context, scraper *scraper.Scraper, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			slog.Debug("Starting scheduled scrape...")
			safeScrape(scraper)
		case <-ctx.Done():
			slog.Debug("Stopping scheduled scrape...")
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
	slog.Debug("Scrape result:", "result", scraperResult)
}
