package main

import (
	"fmt"
	"vinted-watcher/internal/vinted"
)

func main() {
	url := "https://www.vinted.co.uk/catalog?search_text=universal%20works&time=1754856542&material_ids[]=149&material_ids[]=122&page=1"

	params, err := vinted.ParseVintedURL(url)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return
	}

	client := vinted.NewClient("https://www.vinted.co.uk/api/v2")
	items, err := client.GetItems(params)
	if err != nil {
		fmt.Println("Error fetching items:", err)
		return
	}

	for _, item := range items {
		fmt.Printf("Item ID: %d, Title: %s, Price: %s\n", item.ID, item.Title, item.Price.Amount)
	}
}
