package config

import (
	"os"
	"time"
)

type HTTP struct {
	Address      string
	BaseURL      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DB struct {
	DSN string
}

type Log struct {
	Level string
}

type Config struct {
	HTTP HTTP
	DB   DB
	Log  Log
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getdur(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func Load() Config {
	return Config{
		HTTP: HTTP{
			Address:      getenv("HTTP_ADDRESS", ":8080"),
			BaseURL:      getenv("BASE_URL", "http://localhost:8080"),
			ReadTimeout:  getdur("HTTP_READ_TIMEOUT", 5*time.Second),
			WriteTimeout: getdur("HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getdur("HTTP_IDLE_TIMEOUT", 60*time.Second),
		},
		DB: DB{
			DSN: getenv("DATABASE_DSN", "file:tinyurl.db?_journal_mode=WAL&_busy_timeout=5000"),
		},
		Log: Log{
			Level: getenv("LOG_LEVEL", "info"),
		},
	}
}
