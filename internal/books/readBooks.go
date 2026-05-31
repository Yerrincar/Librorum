package books

import (
	db "Librorum/internal/platform/storage/sqlc"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	Db      *pgxpool.Pool
	Queries *db.Queries
	Book    Book
}

type Book struct {
	Title     string `json:"Title"`
	Author    string `json:"Author"`
	CoverPath string `json:"CoverPath"`
}

func (h *Handler) DisplayBooks(w http.ResponseWriter, r *http.Request) {
	appCtx := r.Context()
	setupCtx, cancel := context.WithTimeout(appCtx, 10*time.Second)
	defer cancel()

	limit, offset := ParseLimits(r)
	rows, err := h.Queries.SelectBooks(setupCtx, db.SelectBooksParams{Limit: limit, Offset: offset})
	if err != nil {
		http.Error(w, "The user x doesn't have any books added yet", http.StatusInternalServerError)
	}
	books := make([]*Book, 0)

	for _, row := range rows {
		b := &Book{
			Title:     row.Title,
			Author:    row.Author,
			CoverPath: row.CoverPath,
		}
		books = append(books, b)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(books); err != nil {
		log.Printf("write json response: %v", err)
	}
}

func ParseLimits(r *http.Request) (int32, int32) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit
	return int32(limit), int32(offset)
}
