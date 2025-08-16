package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"vinted-watcher/internal/domain"

	_ "github.com/mattn/go-sqlite3"
)

var now = time.Now()

type DB struct {
	conn *sql.DB
}

func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	db := &DB{conn: conn}

	// Ensure the database is created and ready
	if err := db.conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := db.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}

func (d *DB) CreateSearch(search *domain.SavedSearch) (int, error) {

	searchParamsJSON, err := json.Marshal(*search.SearchParams)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal search params: %w", err)
	}

	result, err := d.conn.Exec(`
        INSERT INTO saved_searches (name, search_params, last_checked, active, created_at, updated_at)
        VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		search.Name, searchParamsJSON, search.LastChecked, search.Active)

	if err != nil {
		return 0, fmt.Errorf("failed to execute insert query: %w", err)
	}

	searchID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return int(searchID), nil
}

func (d *DB) GetSearchByID(id int) (*domain.SavedSearch, error) {
	var search domain.SavedSearch
	var searchParamsJSON string

	err := d.conn.QueryRow(`
        SELECT *
        FROM saved_searches
        WHERE id = ?`, id).Scan(&search.ID, &search.Name, &searchParamsJSON, &search.LastChecked, &search.Active, &search.CreatedAt, &search.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to execute select query: %w", err)
	}

	if err := json.Unmarshal([]byte(searchParamsJSON), &search.SearchParams); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search params: %w", err)
	}

	return &search, nil
}

func (d *DB) GetAllSearches() ([]*domain.SavedSearch, error) {
	rows, err := d.conn.Query(`
        SELECT *
        FROM saved_searches`)
	if err != nil {
		return nil, fmt.Errorf("failed to execute select query: %w", err)
	}
	defer rows.Close()

	var searches []*domain.SavedSearch
	for rows.Next() {
		var search domain.SavedSearch
		var searchParamsJSON string

		if err := rows.Scan(&search.ID, &search.Name, &searchParamsJSON, &search.LastChecked, &search.Active, &search.CreatedAt, &search.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if err := json.Unmarshal([]byte(searchParamsJSON), &search.SearchParams); err != nil {
			return nil, fmt.Errorf("failed to unmarshal search params: %w", err)
		}

		searches = append(searches, &search)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	return searches, nil
}

func (d *DB) IsItemSeen(searchID int, itemID int) (bool, error) {
	var seen bool
	err := d.conn.QueryRow(`
        SELECT EXISTS(
            SELECT 1
            FROM seen_items
            WHERE search_id = ? AND item_id = ?
        )`, searchID, itemID).Scan(&seen)

	if err != nil {
		return false, fmt.Errorf("failed to execute select query: %w", err)
	}

	return seen, nil
}

func (d *DB) MarkItemAsSeen(searchID int, itemID int) error {
	_, err := d.conn.Exec(`
        INSERT INTO seen_items (search_id, item_id)
        VALUES (?, ?)`, searchID, itemID)

	return err
}

func (d *DB) Close() error {
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}

func (db *DB) createTables() error {
	createSearchesTable := `
    CREATE TABLE IF NOT EXISTS saved_searches (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        search_params TEXT NOT NULL,
        last_checked DATETIME,
        active BOOLEAN DEFAULT 1,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`

	createSeenItemsTable := `
    CREATE TABLE IF NOT EXISTS seen_items (
        search_id INTEGER NOT NULL,
        item_id INTEGER NOT NULL,
        seen_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (search_id, item_id),
        FOREIGN KEY (search_id) REFERENCES saved_searches(id) ON DELETE CASCADE
    );`

	if _, err := db.conn.Exec(createSearchesTable); err != nil {
		return err
	}

	if _, err := db.conn.Exec(createSeenItemsTable); err != nil {
		return err
	}

	return nil
}
