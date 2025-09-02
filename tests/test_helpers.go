package tests

import (
	"context"
	"database/sql"
	"log/slog"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"tinyurl/internal/logger"
	"tinyurl/internal/repository"
	"tinyurl/internal/service/link"
	"tinyurl/internal/storage/sqlite"
	sqlrepo "tinyurl/internal/storage/sqlite/repository"
)

func NewTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlite.Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	return db
}

type TestDeps struct {
	DB  *sql.DB
	Rep repository.LinkRepository
	Svc *link.Service
	Log *slog.Logger
}

func NewTestDeps(t *testing.T) *TestDeps {
	t.Helper()
	db := NewTestDB(t)
	rep := sqlrepo.NewLinkRepo(db)
	svc := link.New(rep)
	log := logger.New("error")
	return &TestDeps{DB: db, Rep: rep, Svc: svc, Log: log}
}
