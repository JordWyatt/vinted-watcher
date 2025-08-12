package vinted

import (
	"testing"
	"vinted-watcher/internal/domain"

	"github.com/stretchr/testify/require"
)

func Test_Parse_WithFullVintedURL(t *testing.T) {
	url := "https://www.vinted.co.uk/catalog?search_text=universal%20works%20men&time=1754854403&catalog[]=2051&catalog[]=2052&page=1&size_ids[]=209&size_ids[]=210&brand_ids[]=378695&brand_ids[]=378696&status_ids[]=6&status_ids[]=7&patterns_ids[]=28&patterns_ids[]=29&price_from=0&currency=GBP&price_to=100&order=newest_first"

	expected := &domain.SearchParams{
		SearchText:  "universal works men",
		Time:        1754854403,
		CatalogIDs:  []int{2051, 2052},
		Page:        1,
		SizeIDs:     []int{209, 210},
		BrandIDs:    []int{378695, 378696},
		StatusIDs:   []int{6, 7},
		PatternsIDs: []int{28, 29},
		PriceFrom:   0,
		PriceTo:     100,
		Currency:    "GBP",
	}

	actual, err := ParseVintedURL(url)
	require.NoError(t, err, "should not return an error for a valid URL")

	require.Equal(t, expected, actual, "parsed parameters should match expected values")
}

func Test_Parse_MinimalURL(t *testing.T) {
	url := "https://www.vinted.co.uk/catalog?search_text=universal%20works&time=1754855320"

	expected := &domain.SearchParams{
		SearchText: "universal works",
		Time:       1754855320,
	}

	actual, err := ParseVintedURL(url)
	require.NoError(t, err, "should not return an error for a valid URL")

	require.Equal(t, expected, actual, "parsed parameters should match expected values")
}
