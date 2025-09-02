package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type HTTP struct {
	Address      string        `yaml:"address"`
	BaseURL      string        `yaml:"base_url"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

type DB struct {
	DSN string `yaml:"dsn"`
}

type Log struct {
	Level string `yaml:"level"`
}

type Config struct {
	HTTP HTTP `yaml:"http"`
	DB   DB   `yaml:"db"`
	Log  Log  `yaml:"log"`
}

func Load() (Config, error) {
	cfg := Config{
		HTTP: HTTP{
			Address:      ":8080",
			BaseURL:      "http://localhost:8080",
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		DB: DB{
			DSN: "file:tinyurl.db?_journal_mode=WAL&_busy_timeout=5000",
		},
		Log: Log{
			Level: "info",
		},
	}

	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "config.yaml"
	}

	data, err := os.ReadFile(configFile)
	if err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return cfg, err
		}
	}

	if v := os.Getenv("HTTP_ADDRESS"); v != "" {
		cfg.HTTP.Address = v
	}
	if v := os.Getenv("BASE_URL"); v != "" {
		cfg.HTTP.BaseURL = v
	}
	if v := os.Getenv("HTTP_READ_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.HTTP.ReadTimeout = d
		}
	}
	if v := os.Getenv("HTTP_WRITE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.HTTP.WriteTimeout = d
		}
	}
	if v := os.Getenv("HTTP_IDLE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.HTTP.IdleTimeout = d
		}
	}

	if v := os.Getenv("DATABASE_DSN"); v != "" {
		cfg.DB.DSN = v
	}

	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}

	return cfg, nil
}
