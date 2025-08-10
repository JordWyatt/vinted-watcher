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

	apiURL, err := params.ToApiURL()
	if err != nil {
		fmt.Println("Error generating API URL:", err)
		return
	}

	fmt.Println("Generated API URL:", apiURL)
}
