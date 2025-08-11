package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

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
        original_url TEXT NOT NULL,
        search_params TEXT NOT NULL,
        last_checked DATETIME,
        active BOOLEAN DEFAULT 1,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
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
