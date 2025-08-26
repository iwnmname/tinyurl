package repository

import (
	"context"
	"time"
)

type Link struct {
	ID        int64
	Code      string
	URL       string
	CreatedAt time.Time
}

type LinkRepository interface {
	Create(ctx context.Context, l *Link) error
	GetByCode(ctx context.Context, code string) (*Link, error)
	GetByURL(ctx context.Context, url string) (*Link, error)
	Delete(ctx context.Context, code string) error
}
