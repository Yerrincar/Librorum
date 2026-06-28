package books

import (
	db "Librorum/internal/platform/storage/sqlc"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func (h *Handler) DisplayBooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId, status, err := h.SessionId(ctx, r)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	kind := r.URL.Query().Get("kind")

	limit, offset := ParseLimits(r)
	var rows []db.LibraryItem
	if kind == "" {
		rows, err = h.Queries.SelectBooksByUser(ctx, db.SelectBooksByUserParams{UserID: userId, Limit: limit, Offset: offset})
		if err != nil {
			http.Error(w, "Error trying to select books by user", http.StatusInternalServerError)
		}
	} else {
		rows, err = h.Queries.SelectBooksByUserAndKind(ctx, db.SelectBooksByUserAndKindParams{UserID: userId, Kind: kind, Limit: limit, Offset: offset})
		if err != nil {
			http.Error(w, "Error trying to select books by user and kind", http.StatusInternalServerError)
		}

	}
	books := make([]*Book, 0)

	for _, row := range rows {
		b := &Book{
			Title:      row.Title,
			Author:     row.Author,
			Cover_path: row.CoverPath,
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
		limit = 50
	}

	offset := (page - 1) * limit
	return int32(limit), int32(offset)
}
