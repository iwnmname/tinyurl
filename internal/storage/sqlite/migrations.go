package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func Migrate(ctx context.Context, db *sql.DB) error {
	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		b, err := migrationFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		stmts := strings.Split(string(b), ";")
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		for _, s := range stmts {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			if _, err := tx.ExecContext(ctx, s); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("migration %s failed: %w", name, err)
			}
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
