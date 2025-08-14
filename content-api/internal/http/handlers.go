package http

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"telegramBot/content-api/internal/repo"
)

type Server struct {
	DB *pgxpool.Pool
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/content", s.getContent)
	return mux
}

func (s *Server) getContent(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	lang := r.URL.Query().Get("lang")
	if slug == "" {
		http.Error(w, "missing slug", http.StatusBadRequest)
		return
	}
	if lang != "kz" && lang != "ru" {
		lang = "ru"
	}
	c, err := repo.GetBySlugLang(r.Context(), s.DB, slug, lang)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}
