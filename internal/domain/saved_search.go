package domain

import "time"

type SavedSearch struct {
	ID           int
	Name         string
	OriginalURL  string
	SearchParams *SearchParams
	LastChecked  time.Time
	Active       bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewSavedSearch(searchParams *SearchParams) *SavedSearch {
	return &SavedSearch{
		Name:         searchParams.SearchText,
		SearchParams: searchParams,
		Active:       true,
	}
}
