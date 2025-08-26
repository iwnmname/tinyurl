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
	ExpiresAt *time.Time
	HitCount  int
}

type LinkRepository interface {
	Create(ctx context.Context, l *Link) error
	GetByCode(ctx context.Context, code string) (*Link, error)
	GetByURL(ctx context.Context, url string) (*Link, error)
	IncrementHit(ctx context.Context, code string) error
	Delete(ctx context.Context, code string) error
}
