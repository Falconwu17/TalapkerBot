package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Content struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func GetBySlugLang(ctx context.Context, db *pgxpool.Pool, slug, lang string) (Content, error) {
	var c Content
	err := db.QueryRow(ctx, `select title, body from content where slug=$1 and lang=$2`, slug, lang).
		Scan(&c.Title, &c.Body)
	if err != nil && lang != "ru" {
		err = db.QueryRow(ctx, `select title, body from content where slug=$1 and lang='ru'`, slug).
			Scan(&c.Title, &c.Body)
	}
	return c, err
}
