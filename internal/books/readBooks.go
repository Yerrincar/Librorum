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
			return
		}
	} else {
		rows, err = h.Queries.SelectBooksByUserAndKind(ctx, db.SelectBooksByUserAndKindParams{UserID: userId, Kind: kind, Limit: limit, Offset: offset})
		if err != nil {
			http.Error(w, "Error trying to select books by user and kind", http.StatusInternalServerError)
			return
		}

	}
	books := make([]*Book, 0)

	for _, row := range rows {
		rating, err := row.Rating.Float64Value()
		if err != nil {
			h.Logger.Error("Error trying to parse book rating: "+err.Error(), nil)
		}
		currentChapter, err := row.CurrentChapter.Float64Value()
		if err != nil {
			h.Logger.Error("Error trying to parse current chapter: "+err.Error(), nil)
		}
		totalChapters, err := row.TotalChapters.Float64Value()
		if err != nil {
			h.Logger.Error("Error trying to parse total chapters: "+err.Error(), nil)
		}

		b := &Book{
			ID:               row.ID,
			Title:            row.Title,
			Author:           row.Author,
			Kind:             row.Kind,
			Description:      row.Description,
			Language:         row.Language,
			Genres:           row.Genres,
			Ownership_status: row.OwnershipStatus,
			Reading_status:   row.ReadingStatus,
			Cover_path:       row.CoverPath,
			Notes:            row.Notes,
		}
		if rating.Valid {
			ratingValue := rating.Float64
			b.Rating = &ratingValue
		}
		if currentChapter.Valid {
			currentChapterValue := currentChapter.Float64
			b.Current_chapter = &currentChapterValue
		}
		if totalChapters.Valid {
			totalChaptersValue := totalChapters.Float64
			b.Total_chapters = &totalChaptersValue
		}
		if row.ReadAt.Valid {
			readAt := row.ReadAt.Time
			b.Read_at = &readAt
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
