package storage

import (
	"testing"
	"time"
	"vinted-watcher/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CreateAndGetSearch(t *testing.T) {
	// Initialize the database
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create a sample search
	searchParams := &domain.SearchParams{
		SearchText: "test search",
		Time:       1754856542,
	}
	savedSearch := domain.NewSavedSearch(searchParams)

	before := time.Now().UTC().Truncate(time.Second)

	// Attempt to create the search in the database
	id, err := db.CreateSearch(savedSearch)

	after := time.Now().UTC().Truncate(time.Second)

	if err != nil {
		t.Fatalf("Failed to create search: %v", err)
	}

	assert.Equal(t, 1, id)

	search, err := db.GetSearchByID(id)
	require.NoError(t, err)

	assert.NotNil(t, search)
	assert.Equal(t, 1, search.ID)
	assert.Equal(t, savedSearch.Name, search.Name)
	assert.Equal(t, savedSearch.SearchParams, search.SearchParams)
	assert.Equal(t, savedSearch.LastChecked, search.LastChecked)
	assert.Equal(t, savedSearch.Active, search.Active)
	assert.WithinRange(t, search.CreatedAt, before, after)
	assert.WithinRange(t, search.UpdatedAt, before, after)
}
