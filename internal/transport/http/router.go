package http

import (
	"log/slog"
	"net/http"
)

func NewRouter(h *Handlers, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.Health)
	mux.HandleFunc("POST /shorten", h.Shorten)
	mux.HandleFunc("GET /r/{code}", h.Redirect)
	mux.HandleFunc("GET /stats/{code}", h.Stats)

	// совместимость/запасной путь
	mux.HandleFunc("GET /{code}", h.Redirect)

	return WithMiddlewares(mux, RequestLogger(log))
}
