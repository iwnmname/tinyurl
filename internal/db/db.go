package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"
	"tinyurl/internal/models"
)

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("ошибка при проверке соединения: %w", err)
	}

	_, err = db.Exec(`PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL; PRAGMA foreign_keys=ON;`)
	if err != nil {
		return nil, fmt.Errorf("ошибка при настройке SQLite: %w", err)
	}

	schemaSQL, err := os.ReadFile("db/schema.sql")
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении файла схемы: %w", err)
	}

	_, err = db.Exec(string(schemaSQL))
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении схемы: %w", err)
	}

	return db, nil
}

func InsertLink(db *sql.DB, code, url string, ttlDays int) error {
	var expires interface{}
	if ttlDays > 0 {
		expires = time.Now().AddDate(0, 0, ttlDays)
	}

	_, err := db.Exec("INSERT INTO links (code, url, expires_at) VALUES (?, ?, ?)", code, url, expires)
	if err != nil {
		return fmt.Errorf("ошибка при вставке ссылки: %w", err)
	}

	return nil
}

func IsUniqueError(err error) bool {
	return err != nil && err.Error() != "" && contains(err.Error(), "UNIQUE constraint failed")
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func GetLink(db *sql.DB, code string) (*models.Link, error) {
	var link models.Link
	var expires sql.NullTime

	err := db.QueryRow(`
		SELECT id, code, url, created_at, expires_at, hit_count 
		FROM links 
		WHERE code = ?`, code).Scan(
		&link.ID, &link.Code, &link.URL, &link.CreatedAt, &expires, &link.HitCount)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка при получении ссылки: %w", err)
	}

	if expires.Valid {
		link.ExpiresAt = &expires.Time
	}

	return &link, nil
}

func IncrementHitCount(db *sql.DB, code string) error {
	result, err := db.Exec("UPDATE links SET hit_count = hit_count + 1 WHERE code = ?", code)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении счетчика: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("ссылка не найдена")
	}

	return nil
}
