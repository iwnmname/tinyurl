package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"tinyurl/internal/repository"
)

type LinkRepo struct {
	db *sql.DB
}

func NewLinkRepo(db *sql.DB) *LinkRepo {
	return &LinkRepo{db: db}
}

func (r *LinkRepo) Create(ctx context.Context, l *repository.Link) error {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO links(code, url, created_at) VALUES(?, ?, ?)",
		l.Code, l.URL, time.Now(),
	)
	if err != nil {
		return err
	}
	if id, err := res.LastInsertId(); err == nil {
		l.ID = id
	}
	return nil
}

func (r *LinkRepo) GetByCode(ctx context.Context, code string) (*repository.Link, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, code, url, created_at FROM links WHERE code = ? LIMIT 1", code)
	var l repository.Link
	if err := row.Scan(&l.ID, &l.Code, &l.URL, &l.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}

func (r *LinkRepo) GetByURL(ctx context.Context, url string) (*repository.Link, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, code, url, created_at FROM links WHERE url = ? LIMIT 1", url)
	var l repository.Link
	if err := row.Scan(&l.ID, &l.Code, &l.URL, &l.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}

func (r *LinkRepo) Delete(ctx context.Context, code string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM links WHERE code = ?", code)
	return err
}
