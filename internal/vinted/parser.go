package internal

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type SearchParams struct {
	// Required parameters
	SearchText string

	// Optional parameters
	Time        int64
	CatalogIDs  []int
	Page        int
	SizeIDs     []int
	BrandIDs    []int
	StatusIDs   []int
	PatternsIDs []int
	PriceFrom   float64
	PriceTo     float64
	Currency    string
	Order       string
}

func ParseVintedURL(u string) (*SearchParams, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	params := &SearchParams{}

	if searchText := parsedURL.Query().Get("search_text"); searchText != "" {
		params.SearchText = searchText
	} else {
		return nil, fmt.Errorf("missing required parameter: search_text")
	}

	if currency := parsedURL.Query().Get("currency"); currency != "" {
		params.Currency = currency
	}

	if order := parsedURL.Query().Get("order"); order != "" {
		params.Order = order
	}

	if timeStr := parsedURL.Query().Get("time"); timeStr != "" {
		time, err := strconv.Atoi(timeStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse time: %w", err)
		}
		params.Time = int64(time)
	}

	if catalogIDsString := parsedURL.Query()["catalog[]"]; len(catalogIDsString) > 0 {
		catalogIDs, err := convertCommaSeparatedToIntSlice(catalogIDsString)
		if err != nil {
			return nil, fmt.Errorf("failed to parse catalog: %w", err)
		}
		params.CatalogIDs = catalogIDs
	}

	if sizeIDsString := parsedURL.Query()["size_ids[]"]; len(sizeIDsString) > 0 {
		sizeIDs, err := convertCommaSeparatedToIntSlice(sizeIDsString)
		if err != nil {
			return nil, fmt.Errorf("failed to parse sizeIDs: %w", err)
		}
		params.SizeIDs = sizeIDs
	}

	if brandIDsString := parsedURL.Query()["brand_ids[]"]; len(brandIDsString) > 0 {
		brandIDs, err := convertCommaSeparatedToIntSlice(brandIDsString)
		if err != nil {
			return nil, fmt.Errorf("failed to parse brandIDs: %w", err)
		}
		params.BrandIDs = brandIDs
	}

	if statusIDsString := parsedURL.Query()["status_ids[]"]; len(statusIDsString) > 0 {
		statusIDs, err := convertCommaSeparatedToIntSlice(statusIDsString)
		if err != nil {
			return nil, fmt.Errorf("failed to parse statusIDs: %w", err)
		}
		params.StatusIDs = statusIDs
	}

	if patternsIDsString := parsedURL.Query()["patterns_ids[]"]; len(patternsIDsString) > 0 {
		patternsIDs, err := convertCommaSeparatedToIntSlice(patternsIDsString)
		if err != nil {
			return nil, fmt.Errorf("failed to parse patternsIDs: %w", err)
		}
		params.PatternsIDs = patternsIDs
	}

	if pageStr := parsedURL.Query().Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse page: %w", err)
		}
		params.Page = page
	}

	if priceFromStr := parsedURL.Query().Get("price_from"); priceFromStr != "" {
		priceFrom, err := strconv.ParseFloat(priceFromStr, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse price_from: %w", err)
		}
		params.PriceFrom = priceFrom
	}

	if priceToStr := parsedURL.Query().Get("price_to"); priceToStr != "" {
		priceTo, err := strconv.ParseFloat(priceToStr, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse price_to: %w", err)
		}
		params.PriceTo = priceTo
	}

	return params, nil
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

	if s.Order != "" {
		values.Set("order", s.Order)
	}

	if s.Time != 0 {
		values.Set("time", fmt.Sprintf("%d", s.Time))
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

	return fmt.Sprintf("%s?%s", baseURL, values.Encode()), nil
}

func convertCommaSeparatedToIntSlice(input []string) ([]int, error) {
	var result []int
	for _, s := range input {
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("failed to convert %q to int: %w", s, err)
		}
		result = append(result, i)
	}
	return result, nil
}
