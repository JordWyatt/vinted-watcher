package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Parse_WithFullVintedURL(t *testing.T) {
	url := "https://www.vinted.co.uk/catalog?search_text=universal%20works%20men&time=1754854403&catalog[]=2051&page=1&size_ids[]=209&brand_ids[]=378695&status_ids[]=6&patterns_ids[]=28&price_from=0&currency=GBP&price_to=100&order=newest_first"

	expected := &SearchParams{
		SearchText:  "universal works men",
		Time:        1754854403,
		CatalogIDs:  []int{2051},
		Page:        1,
		SizeIDs:     []int{209},
		BrandIDs:    []int{378695},
		StatusIDs:   []int{6},
		PatternsIDs: []int{28},
		PriceFrom:   0,
		PriceTo:     100,
		Currency:    "GBP",
		Order:       "newest_first",
	}

	actual, err := ParseVintedURL(url)
	require.NoError(t, err, "should not return an error for a valid URL")

	require.Equal(t, expected, actual, "parsed parameters should match expected values")
}

func Test_Parse_MinimalURL(t *testing.T) {
	url := "https://www.vinted.co.uk/catalog?search_text=universal%20works&time=1754855320"

	expected := &SearchParams{
		SearchText: "universal works",
		Time:       1754855320,
	}

	actual, err := ParseVintedURL(url)
	require.NoError(t, err, "should not return an error for a valid URL")

	require.Equal(t, expected, actual, "parsed parameters should match expected values")
}
