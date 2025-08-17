package domain

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ToApiURL_WithCompleteSearchTerms(t *testing.T) {
	params := &SearchParams{
		SearchText: "universal works men",

		CatalogIDs:  []int{2051, 2052},
		Page:        1,
		SizeIDs:     []int{209, 210},
		BrandIDs:    []int{378695, 378696},
		StatusIDs:   []int{6, 7},
		PatternsIDs: []int{28, 29},
		PriceFrom:   1,
		PriceTo:     100,
		Currency:    "GBP",
	}

	expectedURL := "https://www.vinted.co.uk/api/v2/catalog/items?brand_ids[]=378695&brand_ids[]=378696&catalog_ids[]=2051&catalog_ids[]=2052&currency=GBP&order=newest_first&page=1&patterns_ids[]=28&patterns_ids[]=29&price_from=1.00&price_to=100.00&search_text=universal+works+men&size_ids[]=209&size_ids[]=210&status_ids[]=6&status_ids[]=7&time=1754854403"
	actualURL, err := params.ToApiURL()
	escapedActualURL, err := url.QueryUnescape(actualURL)

	require.NoError(t, err, "should not return an error for valid parameters")

	assert.Equal(t, expectedURL, escapedActualURL, "generated API URL should match expected URL")
}

func Test_ToApiURL_WithMinimalSearchTerms(t *testing.T) {
	params := &SearchParams{
		SearchText: "universal works",
	}

	expectedURL := "https://www.vinted.co.uk/api/v2/catalog/items?order=newest_first&search_text=universal+works&time=1754855320"
	actualURL, err := params.ToApiURL()
	escapedActualURL, err := url.QueryUnescape(actualURL)

	require.NoError(t, err, "should not return an error for valid parameters")

	assert.Equal(t, expectedURL, escapedActualURL, "generated API URL should match expected URL")
}
