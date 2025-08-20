package tests

import (
	"database/sql"
	"testing"

	"tinyurl/internal/db"

	_ "modernc.org/sqlite"
)

func TestInsertAndGetLink(t *testing.T) {
	database, err := sql.Open("sqlite", "file:test_db.db?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	_, err = database.Exec(`
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
		t.Fatalf("Failed to create schema: %v", err)
	}

	testCases := []struct {
		name    string
		code    string
		url     string
		ttlDays int
	}{
		{"No TTL", "test1", "https://example.com", 0},
		{"With TTL", "test2", "https://example.org", 7},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := db.InsertLink(database, tc.code, tc.url, tc.ttlDays)
			if err != nil {
				t.Fatalf("InsertLink failed: %v", err)
			}

			link, err := db.GetLink(database, tc.code)
			if err != nil {
				t.Fatalf("GetLink failed: %v", err)
			}

			if link == nil {
				t.Fatal("GetLink returned nil")
			}
			if link.Code != tc.code {
				t.Errorf("Expected code %s, got %s", tc.code, link.Code)
			}
			if link.URL != tc.url {
				t.Errorf("Expected URL %s, got %s", tc.url, link.URL)
			}
			if tc.ttlDays > 0 {
				if link.ExpiresAt == nil {
					t.Error("Expected non-nil ExpiresAt")
				}
			} else {
				if link.ExpiresAt != nil {
					t.Errorf("Expected nil ExpiresAt, got %v", *link.ExpiresAt)
				}
			}
		})
	}
}

func TestIncrementHitCount(t *testing.T) {
	database, err := sql.Open("sqlite", "file:test_db_inc.db?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	_, err = database.Exec(`
	CREATE TABLE IF NOT EXISTS links (
	  id          INTEGER PRIMARY KEY AUTOINCREMENT,
	  code        TEXT    NOT NULL UNIQUE,
	  url         TEXT    NOT NULL,
	  created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	  expires_at  TIMESTAMP NULL,
	  hit_count   INTEGER NOT NULL DEFAULT 0
	);
	`)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	code := "increment_test"
	url := "https://example.com"

	err = db.InsertLink(database, code, url, 0)
	if err != nil {
		t.Fatalf("InsertLink failed: %v", err)
	}

	link, err := db.GetLink(database, code)
	if err != nil {
		t.Fatalf("GetLink failed: %v", err)
	}
	if link.HitCount != 0 {
		t.Errorf("Initial hit count should be 0, got %d", link.HitCount)
	}

	err = db.IncrementHitCount(database, code)
	if err != nil {
		t.Fatalf("IncrementHitCount failed: %v", err)
	}

	link, err = db.GetLink(database, code)
	if err != nil {
		t.Fatalf("GetLink failed after increment: %v", err)
	}
	if link.HitCount != 1 {
		t.Errorf("Hit count should be 1 after increment, got %d", link.HitCount)
	}
}
