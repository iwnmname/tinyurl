package main

import (
	"fmt"
	"log"
	_ "modernc.org/sqlite"
	"net/http"

	"tinyurl/internal/db"
	"tinyurl/internal/handlers"
)

func main() {
	fmt.Println("Запуск TinyURL...")

	database, err := db.InitDB("file:tinyurl.db?cache=shared&mode=rwc&_fk=1")
	if err != nil {
		log.Fatal("Ошибка при инициализации базы данных:", err)
	}
	defer database.Close()

	server := handlers.NewServer(database)

	http.HandleFunc("/shorten", server.ShortenHandler)
	http.HandleFunc("/r/", server.RedirectHandler)
	http.HandleFunc("/stats/", server.StatsHandler)

	fmt.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
