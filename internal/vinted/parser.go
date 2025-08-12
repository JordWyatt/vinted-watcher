package vinted

import (
	"fmt"
	"net/url"
	"strconv"
	"vinted-watcher/internal/domain"
)

func ParseVintedURL(u string) (*domain.SearchParams, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	params := &domain.SearchParams{}

	if searchText := parsedURL.Query().Get("search_text"); searchText != "" {
		params.SearchText = searchText
	} else {
		return nil, fmt.Errorf("missing required parameter: search_text")
	}

	if currency := parsedURL.Query().Get("currency"); currency != "" {
		params.Currency = currency
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
