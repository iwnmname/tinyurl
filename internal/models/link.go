package models

import "time"

type Link struct {
	ID        int64
	Code      string
	URL       string
	CreatedAt time.Time
	ExpiresAt *time.Time
	HitCount  int64
}

type ShortenRequest struct {
	URL     string `json:"url"`
	Alias   string `json:"alias,omitempty"`
	TTLDays int    `json:"ttl_days,omitempty"`
}

type ShortenResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}

type StatsResponse struct {
	URL       string     `json:"url"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	HitCount  int64      `json:"hit_count"`
}
