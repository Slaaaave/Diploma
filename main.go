package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Slaaaave/Diploma/pkg/api"
	"github.com/Slaaaave/Diploma/pkg/db"
)

const (
	defaultPort   = "7540"
	defaultDBFile = "scheduler.db"
	webDir        = "./web"
)

func main() {
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = defaultPort
	}

	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = defaultDBFile
	}

	if err := db.Init(dbFile); err != nil {
		log.Fatalf("ошибка инициализации базы данных: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(webDir)))
	api.Init(mux)

	log.Printf("сервер запущен на порту %s (база данных: %s)", port, dbFile)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("ошибка запуска сервера: %v", err)
	}
}
