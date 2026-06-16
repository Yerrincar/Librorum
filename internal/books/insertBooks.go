package books

import (
	db "Librorum/internal/platform/storage/sqlc"
	"Librorum/internal/users"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) InsertEpubBooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, status, err := h.SessionId(ctx, r)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	err = r.ParseMultipartForm(20 << 20)
	if err != nil {
		http.Error(w, "Error parsing multipart form", http.StatusBadRequest)
		return
	}
	kind := r.FormValue("kind")
	if kind == "" {
		kind = "book"
	}
	ratingForm := strings.TrimSpace(r.FormValue("rating"))

	var rating pgtype.Numeric
	if ratingForm == "" {
		rating = pgtype.Numeric{Valid: false}
	} else if err := rating.Scan(ratingForm); err != nil {
		http.Error(w, "invalid rating", http.StatusBadRequest)
		return
	}
	ownership_status := r.FormValue("ownership_status")
	if ownership_status == "" {
		ownership_status = "none"
	}
	reading_status := r.FormValue("reading_status")
	if reading_status == "" {
		reading_status = "unread"
	}
	publication_status := r.FormValue("publication_status")
	if publication_status == "" {
		publication_status = "unknown"
	}

	current_chapter := r.FormValue("current_chapter")
	var currentChapter pgtype.Numeric
	if current_chapter == "" {
		currentChapter = pgtype.Numeric{Valid: false}
	} else if err := currentChapter.Scan(current_chapter); err != nil {
		http.Error(w, "invalid current chapter", http.StatusBadRequest)
		return
	}
	read_at := strings.TrimSpace(r.FormValue("read_at"))
	var readAt pgtype.Timestamptz
	if read_at == "" {
		readAt = pgtype.Timestamptz{Valid: false}
	} else {
		parsed, err := time.Parse(time.RFC3339, read_at)
		if err != nil {
			parsed, err = time.ParseInLocation("2006-01-02T15:04", read_at, time.Local)
		}
		if err != nil {
			h.Logger.Error("Invalid read_at value: "+err.Error(), map[string]string{"read_at": read_at})
			http.Error(w, "invalid read at value", http.StatusBadRequest)
			return
		}
		readAt = pgtype.Timestamptz{Time: parsed, Valid: true}
	}

	notes := r.FormValue("notes")

	epubPath, err := h.FileUploader(r)
	if err != nil {
		http.Error(w, "Error trying to upload .epub file", http.StatusBadRequest)
		return
	}

	epubMetadata, err := ExtractEPUB(epubPath)
	if err != nil {
		http.Error(w, "Error trying to extract metadata from .epub file", http.StatusBadRequest)
		return
	}

	coverPath, err := h.Manager.Process(ctx, epubMetadata)
	if err != nil {
		h.Logger.Error("Error trying to get cover path", nil)
	}

	response, err := h.Queries.InsertBook(ctx, db.InsertBookParams{
		UserID:            userId,
		Title:             epubMetadata.Metadata.Title,
		Author:            epubMetadata.Metadata.Author,
		Description:       epubMetadata.Metadata.Description,
		Genres:            epubMetadata.Metadata.Genres,
		Language:          epubMetadata.Metadata.Language,
		PublicationYear:   epubMetadata.Metadata.PublicationYear,
		TotalChapters:     epubMetadata.Metadata.TotalChapters,
		PublicationStatus: publication_status,
		CoverPath:         coverPath,
		Kind:              kind,
		Rating:            rating,
		OwnershipStatus:   ownership_status,
		ReadingStatus:     reading_status,
		CurrentChapter:    currentChapter,
		ReadAt:            readAt,
		Notes:             notes,
	})
	if err != nil {
		h.Logger.Error("Error trying to insert book in the database: "+err.Error(), nil)
		http.Error(w, "Error trying to insert book in the database", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) SessionId(ctx context.Context, r *http.Request) (int64, int, error) {
	sessionToken, err := users.SessionTokenFromRequest(r)
	if err != nil {
		return 0, http.StatusUnauthorized, err
	}
	session, err := h.Queries.FindSessionByTokenHash(ctx, users.HashSessionToken(sessionToken))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, http.StatusUnauthorized, err
		}
		h.Logger.Error("Error trying to find session: "+err.Error(), nil)
		return 0, http.StatusInternalServerError, err
	}

	return session.UserID, http.StatusOK, nil
}
