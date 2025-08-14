package main

import (
	"log"
	"net/http"
	"telegramBot/content-api/internal/db"
	h "telegramBot/content-api/internal/http"
)

func main() {
	pool := db.MustPool()
	defer pool.Close()
	s := &h.Server{DB: pool}
	log.Println("content-api listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", s.Routes()))
}
