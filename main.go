package main

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
	_ "vinted-watcher/internal/logger"
	"vinted-watcher/internal/scraper"
	"vinted-watcher/internal/server"
	"vinted-watcher/internal/storage"
	"vinted-watcher/internal/vinted"
)

// Test code - will eventually become server entrypoint
func main() {
	db, err := storage.NewDB("vinted.db")
	if err != nil {
		slog.Error("Error initializing database:", "error", err)
		return
	}

	wg := sync.WaitGroup{}

	go func() {
		wg.Add(1)
		defer wg.Done()
		server := server.NewServer(db)
		if err := server.Start(); err != nil {
			slog.Error("Error starting server:", "error", err)
		}
	}()
	vintedClient := vinted.NewClient("https://www.vinted.com")

	scraper := scraper.NewScraper(vintedClient, db, scraper.ScraperConfig{
		LookbackPeriod: 24 * time.Hour,
	})

	scraperResult, err := scraper.Scrape()
	if err != nil {
		slog.Error("Error scraping:", "error", err)
		return
	}

	fmt.Printf("Scraper result: %+v\n", scraperResult)

	wg.Wait()
}
