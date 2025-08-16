package storage

import (
	"vinted-watcher/internal/domain"
)

type SearchStorage interface {
	// Search CRUD operations
	CreateSearch(search *domain.SavedSearch) error
	GetSearchByID(id int) (*domain.SavedSearch, error)
	GetAllSearches() ([]domain.SavedSearch, error)
	// UpdateSearch(search *domain.SavedSearch) error
	// DeleteSearch(id int) error

	// Search status management
	// UpdateLastChecked(searchID int) error
	// SetSearchActive(searchID int, active bool) error

	// Item tracking
	MarkItemAsSeen(searchID int, vintedItemID int64) error
	IsItemSeen(searchID int, itemID int64) (bool, error)
	// GetUnseenItems(searchID int, items []vinted.Item) ([]vinted.Item, error)

	// Connection management
	Close() error
}
