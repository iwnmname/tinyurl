package tests

import (
	"database/sql"
	"fmt"
)

// go:coverage ignore
func initTestDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "file:test.db?mode=memory&cache=shared")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	_, err = db.Exec(`PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL; PRAGMA foreign_keys=ON;`)
	if err != nil {
		return nil, fmt.Errorf("failed to configure SQLite: %w", err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS links (
	  id          INTEGER PRIMARY KEY AUTOINCREMENT,
	  code        TEXT    NOT NULL UNIQUE,
	  url         TEXT    NOT NULL,
	  created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	  expires_at  TIMESTAMP NULL,
	  hit_count   INTEGER NOT NULL DEFAULT 0
	);
	CREATE INDEX IF NOT EXISTS idx_links_expires_at ON links (expires_at);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return db, nil
}
