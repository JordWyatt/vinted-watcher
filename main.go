package main

import (
	"fmt"
	"vinted-watcher/internal/domain"
	"vinted-watcher/internal/storage"
	"vinted-watcher/internal/vinted"
)

// Test code - will eventually become server entrypoint
func main() {
	url := "https://www.vinted.co.uk/catalog?search_text=universal%20works&time=1754856542&material_ids[]=149&material_ids[]=122&page=1"

	db, err := storage.NewDB("vinted.db")
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}
	defer db.Close()

	params, err := vinted.ParseVintedURL(url)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return
	}

	savedSearch := domain.NewSavedSearch(params)

	if err := db.CreateSearch(savedSearch); err != nil {
		fmt.Println("Error saving search:", err)
		return
	}

}
