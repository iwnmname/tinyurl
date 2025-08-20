package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"tinyurl/internal/db"
	"tinyurl/internal/models"
	"tinyurl/internal/utils"
)

type Server struct {
	DB *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{DB: db}
}

func (s *Server) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Только метод POST разрешен", http.StatusMethodNotAllowed)
		return
	}

	var req models.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL обязателен", http.StatusBadRequest)
		return
	}

	code := req.Alias
	var err error

	if code == "" {
		for tries := 0; tries < 5; tries++ {
			code = utils.GenerateRandomCode(6)
			if err = db.InsertLink(s.DB, code, req.URL, req.TTLDays); err == nil {
				break
			}
			if !db.IsUniqueError(err) {
				http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
				return
			}
		}
		if err != nil {
			http.Error(w, "Не удалось создать уникальный код, попробуйте снова", http.StatusInternalServerError)
			return
		}
	} else {
		if err = db.InsertLink(s.DB, code, req.URL, req.TTLDays); err != nil {
			if db.IsUniqueError(err) {
				http.Error(w, "Этот алиас уже занят", http.StatusConflict)
				return
			}
			http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
			return
		}
	}

	host := utils.GetHost(r)
	resp := models.ShortenResponse{
		Code:     code,
		ShortURL: fmt.Sprintf("%s/r/%s", host, code),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[len("/r/"):]
	if code == "" {
		http.NotFound(w, r)
		return
	}

	link, err := db.GetLink(s.DB, code)
	if err != nil {
		http.Error(w, "Ошибка при получении ссылки", http.StatusInternalServerError)
		return
	}

	if link == nil {
		http.NotFound(w, r)
		return
	}

	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		http.NotFound(w, r)
		return
	}

	go func() {
		if err := db.IncrementHitCount(s.DB, code); err != nil {
			log.Printf("Ошибка при увеличении счетчика для %s: %v", code, err)
		}
	}()

	http.Redirect(w, r, link.URL, http.StatusFound)
}

func (s *Server) StatsHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[len("/stats/"):]
	if code == "" {
		http.NotFound(w, r)
		return
	}

	link, err := db.GetLink(s.DB, code)
	if err != nil {
		http.Error(w, "Ошибка при получении статистики", http.StatusInternalServerError)
		return
	}

	if link == nil {
		http.NotFound(w, r)
		return
	}

	stats := models.StatsResponse{
		URL:       link.URL,
		CreatedAt: link.CreatedAt,
		ExpiresAt: link.ExpiresAt,
		HitCount:  link.HitCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
