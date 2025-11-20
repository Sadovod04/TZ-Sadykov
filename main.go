package main

import (
	handlers "TZ/handl"
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "user=postgres password=password dbname=postgres sslmode=disable host=localhost port=5432"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS numbers (
			id SERIAL PRIMARY KEY,
			value INTEGER NOT NULL
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	handler := handlers.NewHandler(db)

	http.HandleFunc("/numbers", handler.HandleNumbers)
	http.HandleFunc("/clear", handler.HandleClean)
	log.Println("Server starting on :9091")
	log.Fatal(http.ListenAndServe(":9091", nil))
}
