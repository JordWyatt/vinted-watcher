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
	db := setupTestDB(t)
	defer db.Close()

	// Create a sample search
	searchParams := &domain.SearchParams{
		SearchText: "test search",
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

func Test_GetAllSearches(t *testing.T) {
	// Initialize the database
	db := setupTestDB(t)
	defer db.Close()

	expectedSearches := []*domain.SavedSearch{
		{
			ID:   1,
			Name: "test search",
			SearchParams: &domain.SearchParams{
				SearchText: "test search",
			},
			Active: true,
		},
		{
			ID:   2,
			Name: "another test search",
			SearchParams: &domain.SearchParams{
				SearchText: "another test search",
			},
			Active: true,
		},
	}

	for _, search := range expectedSearches {
		_, err := db.CreateSearch(search)
		require.NoError(t, err)
	}

	actualSearches, err := db.GetAllSearches()
	require.NoError(t, err)

	assert.Len(t, actualSearches, 2)
	assert.Equal(t, expectedSearches[0].ID, actualSearches[0].ID)
	assert.Equal(t, expectedSearches[1].ID, actualSearches[1].ID)
}

func setupTestDB(t *testing.T) *DB {
	t.Helper()

	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	return db
}

func Test_MarkItemAsSeenAndIsItemSeen(t *testing.T) {
	// Initialize the database
	db := setupTestDB(t)
	defer db.Close()

	searchParams := &domain.SearchParams{
		SearchText: "test search",
	}
	savedSearch := domain.NewSavedSearch(searchParams)

	searchID, err := db.CreateSearch(savedSearch)
	require.NoError(t, err)

	itemID := 12345
	err = db.MarkItemAsSeen(searchID, itemID)
	require.NoError(t, err)

	isSeen, err := db.IsItemSeen(searchID, itemID)
	require.NoError(t, err)
	assert.True(t, isSeen)

	// Check for a different item ID to ensure it's not seen
	isSeen, err = db.IsItemSeen(searchID, 1234)
	require.NoError(t, err)
	assert.False(t, isSeen)
}
