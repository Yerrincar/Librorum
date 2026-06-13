package books

import (
	db "Librorum/internal/platform/storage/sqlc"
	"Librorum/internal/users"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) InsertEpubBooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, err := h.SessionId(ctx, w, r)
	if err != nil {
		http.Error(w, "Error trying to extract userID from the sessio", http.StatusBadRequest)
		return
	}

	epubMetadata, err := ExtractEPUB("")
	if err != nil {
		http.Error(w, "Error trying to extract metadata from .epub file", http.StatusBadRequest)
		return
	}
	coverPath, err := h.Manager.Process(ctx, epubMetadata)

	var input struct {
		Kind             string         `json:"Kind"`
		Rating           pgtype.Numeric `json:"Rating"`
		Ownership_status string         `json:"Ownership_status"`
		Reading_status   string         `json:"Reading_status"`
		Current_chapter  pgtype.Numeric `json:"Current_chapter"`
		Read_at          time.Time      `json:"Read_at"`
		Notes            string         `json:"Notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	_, err = h.Queries.InsertBook(ctx, db.InsertBookParams{
		UserID:          userId,
		Title:           epubMetadata.Metadata.Title,
		Author:          epubMetadata.Metadata.Author,
		Description:     epubMetadata.Metadata.Description,
		Genres:          epubMetadata.Metadata.Genres,
		Language:        epubMetadata.Metadata.Language,
		PublicationYear: epubMetadata.Metadata.PublicationYear,
		TotalChapters:   epubMetadata.Metadata.TotalChapters,
		CoverPath:       coverPath,
		Kind:            input.Kind,
		Rating:          input.Rating,
		OwnershipStatus: input.Ownership_status,
		ReadingStatus:   input.Reading_status,
		CurrentChapter:  input.Current_chapter,
		Notes:           input.Notes,
	})
	if err != nil {
		h.Logger.Error("Error trying to insert book in the database: "+err.Error(), nil)
	}
}

func (h *Handler) SessionId(ctx context.Context, w http.ResponseWriter, r *http.Request) (int64, error) {
	sessionToken, err := users.SessionTokenFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized status", http.StatusUnauthorized)
		return 0, err
	}
	session, err := h.Queries.FindSessionByTokenHash(ctx, users.HashSessionToken(sessionToken))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Unauthorized status", http.StatusUnauthorized)
			return 0, err
		}
		h.Logger.Error("Error trying to find session: "+err.Error(), nil)
		http.Error(w, "There was a problem and we couldn't fulfill your request", http.StatusInternalServerError)
		return 0, err
	}

	return session.UserID, nil
}
