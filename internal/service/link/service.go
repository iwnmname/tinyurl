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

type Service struct {
	repo repository.LinkRepository
}

func New(repo repository.LinkRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Shorten(ctx context.Context, url string) (string, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return "", errors.New("empty url")
	}

	if ex, _ := s.repo.GetByURL(ctx, url); ex != nil {
		return ex.Code, nil
	}

	for i := 0; i < 6; i++ {
		code := genCode(7 + i%2)
		if got, _ := s.repo.GetByCode(ctx, code); got != nil {
			continue
		}
		err := s.repo.Create(ctx, &repository.Link{
			Code:      code,
			URL:       url,
			CreatedAt: time.Now(),
		})
		if err != nil {
			continue
		}
		return code, nil
	}
	return "", errors.New("cannot generate unique code")
}

func (s *Service) Resolve(ctx context.Context, code string) (string, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "", errors.New("empty code")
	}
	l, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return "", err
	}
	if l == nil {
		return "", errors.New("not found")
	}
	return l.URL, nil
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
