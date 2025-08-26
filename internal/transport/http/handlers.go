package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"tinyurl/internal/service/link"
)

type Handlers struct {
	svc     *link.Service
	log     *slog.Logger
	baseURL string
}

func NewHandlers(svc *link.Service, log *slog.Logger, baseURL string) *Handlers {
	return &Handlers{svc: svc, log: log, baseURL: strings.TrimRight(baseURL, "/")}
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (h *Handlers) Shorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	code, expiresAt, err := h.svc.Shorten(r.Context(), req.URL, req.Alias, req.TTLDays)
	if err != nil {
		switch {
		case errors.Is(err, link.ErrAliasBusy):
			http.Error(w, "alias is already in use", http.StatusConflict)
		default:
			http.Error(w, "cannot shorten", http.StatusBadRequest)
		}
		h.log.Warn("shorten failed", "err", err)
		return
	}
	short := h.baseURL + "/r/" + code
	resp := ShortenResponse{Code: code, ShortURL: short}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)

	if expiresAt != nil {
		h.log.Info("shortened", "code", code, "url", req.URL, "expires_at", expiresAt.Format(time.RFC3339))
	} else {
		h.log.Info("shortened", "code", code, "url", req.URL)
	}
}

func (h *Handlers) Redirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		code = strings.TrimPrefix(r.URL.Path, "/r/")
		code = strings.TrimPrefix(code, "/")
	}
	url, err := h.svc.Resolve(r.Context(), code)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
	h.log.Info("redirect", "code", code, "to", url)
}

func (h *Handlers) Stats(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		code = strings.TrimPrefix(r.URL.Path, "/stats/")
	}
	l, err := h.svc.Stats(r.Context(), code)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	var expiresStr *string
	if l.ExpiresAt != nil {
		s := l.ExpiresAt.UTC().Format(time.RFC3339)
		expiresStr = &s
	}

	resp := StatsResponse{
		URL:       l.URL,
		CreatedAt: l.CreatedAt.UTC().Format(time.RFC3339),
		ExpiresAt: expiresStr,
		HitCount:  l.HitCount,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
