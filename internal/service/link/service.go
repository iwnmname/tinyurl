package link

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"tinyurl/internal/repository"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrAliasBusy = errors.New("alias is already in use")
	ErrExpired   = errors.New("link expired")
)

type Service struct {
	repo repository.LinkRepository
}

func New(repo repository.LinkRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Shorten(ctx context.Context, url, alias string, ttlDays int) (string, *time.Time, error) {
	url = strings.TrimSpace(url)
	alias = strings.TrimSpace(alias)
	if url == "" {
		return "", nil, errors.New("empty url")
	}

	if alias != "" {
		ex, err := s.repo.GetByCode(ctx, alias)
		if err != nil {
			return "", nil, err
		}
		if ex != nil {
			if !isExpired(ex) {
				return "", nil, ErrAliasBusy
			}
			if err := s.repo.SoftDelete(ctx, alias); err != nil {
				return "", nil, err
			}
		}
	}

	var expiresAt *time.Time
	if ttlDays > 0 {
		t := time.Now().Add(time.Duration(ttlDays) * 24 * time.Hour)
		expiresAt = &t
	}

	if alias == "" {
		if ex, _ := s.repo.GetByURL(ctx, url); ex != nil && !isExpired(ex) {
			return ex.Code, ex.ExpiresAt, nil
		}
	}

	code := alias
	if code == "" {
		var err error
		code, err = s.generateCode(ctx)
		if err != nil {
			return "", nil, err
		}
	}

	l := &repository.Link{
		Code:      code,
		URL:       url,
		ExpiresAt: expiresAt,
	}
	if err := s.repo.Create(ctx, l); err != nil {
		return "", nil, err
	}
	return code, expiresAt, nil
}

func (s *Service) generateCode(ctx context.Context) (string, error) {
	for i := 0; i < 6; i++ {
		cand := genCode(7 + i%2)

		got, _ := s.repo.GetByCode(ctx, cand)
		if got == nil || isExpired(got) {
			return cand, nil
		}
	}
	return "", errors.New("cannot generate code")
}

func (s *Service) Resolve(ctx context.Context, code string) (string, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "", ErrNotFound
	}

	l, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return "", err
	}
	if l == nil {
		return "", ErrNotFound
	}
	if isExpired(l) {
		return "", ErrExpired
	}
	_ = s.repo.IncrementHit(ctx, code)
	return l.URL, nil
}

func (s *Service) Stats(ctx context.Context, code string) (*repository.Link, error) {
	l, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if l == nil {
		return nil, ErrNotFound
	}
	if isExpired(l) {
		return nil, ErrExpired
	}
	return l, nil
}

func isExpired(l *repository.Link) bool {
	return l.ExpiresAt != nil && time.Now().After(*l.ExpiresAt)
}

func (s *Service) Delete(ctx context.Context, code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return ErrNotFound
	}
	l, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return err
	}
	if l == nil {
		return ErrNotFound
	}
	if isExpired(l) {
		return ErrExpired
	}
	return s.repo.SoftDelete(ctx, code)
}

func genCode(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	s := base64.RawURLEncoding.EncodeToString(b)
	if len(s) > n {
		s = s[:n]
	}
	return s
}
