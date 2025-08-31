package domain

import (
	"fmt"
	"net/url"
	"strings"
)

type SearchParams struct {
	// Required parameters
	SearchText string

	// Optional parameters
	CatalogIDs  []int
	Page        int
	SizeIDs     []int
	BrandIDs    []int
	StatusIDs   []int
	PatternsIDs []int
	PriceFrom   float64
	PriceTo     float64
	Currency    string
}

// https://www.vinted.co.uk/api/v2/catalog/items?page=1&per_page=96&time=1754854403&gen_session_id=true&search_text=universal+works+men&catalog_ids=2051&price_from=0&price_to=100&currency=GBP&order=newest_first&size_ids=209&brand_ids=378695&status_ids=6&patterns_ids=28
func (s *SearchParams) ToApiURL() (string, error) {
	baseURL := "https://www.vinted.co.uk/api/v2/catalog/items"
	values := url.Values{}

	if s.SearchText == "" {
		return "", fmt.Errorf("missing required parameter: search_text")
	}

	values.Set("search_text", strings.ReplaceAll(s.SearchText, " ", "+"))

	if s.Currency != "" {
		values.Set("currency", s.Currency)
	}

	if s.Currency != "" {
		values.Set("currency", s.Currency)
	}

	if s.Page != 0 {
		values.Set("page", fmt.Sprintf("%d", s.Page))
	}

	if s.PriceFrom != 0 {
		values.Set("price_from", fmt.Sprintf("%.2f", s.PriceFrom))
	}

	if s.PriceTo != 0 {
		values.Set("price_to", fmt.Sprintf("%.2f", s.PriceTo))
	}

	if len(s.CatalogIDs) > 0 {
		for _, id := range s.CatalogIDs {
			values.Add("catalog_ids[]", fmt.Sprintf("%d", id))
		}
	}
	if len(s.SizeIDs) > 0 {
		for _, id := range s.SizeIDs {
			values.Add("size_ids[]", fmt.Sprintf("%d", id))
		}
	}
	if len(s.BrandIDs) > 0 {
		for _, id := range s.BrandIDs {
			values.Add("brand_ids[]", fmt.Sprintf("%d", id))
		}
	}
	if len(s.StatusIDs) > 0 {
		for _, id := range s.StatusIDs {
			values.Add("status_ids[]", fmt.Sprintf("%d", id))
		}
	}
	if len(s.PatternsIDs) > 0 {
		for _, id := range s.PatternsIDs {
			values.Add("patterns_ids[]", fmt.Sprintf("%d", id))
		}
	}

	values.Add("order", "newest_first")

	encoded := fmt.Sprintf("%s?%s", baseURL, values.Encode())
	// hacky - but the vinted API returns different results if you use the encoded form (often less)
	encoded = strings.ReplaceAll(encoded, "%2B", "+")
	return encoded, nil
}

// TODO: Add ToWebURL method
