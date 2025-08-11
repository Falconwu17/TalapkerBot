package dbConn

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func ConnectToDB() {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Fatal("DATABASE_URL не задан")
	}

	var err error
	for i := 0; i < 20; i++ { // ждём БД до ~20*500мс
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		Pool, err = pgxpool.New(ctx, url)
		cancel()
		if err == nil {
			if ping := Ping(); ping == nil {
				log.Println("Postgres connected")
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	log.Fatalf("не удалось подключиться к Postgres: %v", err)
}

func Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return Pool.Ping(ctx)
}

func Fetch(slug, lang string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var title, body string
	err := Pool.QueryRow(ctx, `SELECT title, body FROM content WHERE slug=$1 AND lang=$2`, slug, lang).Scan(&title, &body)
	if err == nil {
		return title, body, nil
	}
	// фолбэк на русскую версию
	err2 := Pool.QueryRow(ctx, `SELECT title, body FROM content WHERE slug=$1 AND lang='ru'`, slug).Scan(&title, &body)
	return title, body, err2
}
