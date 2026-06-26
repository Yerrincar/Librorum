package books

import (
	db "Librorum/internal/platform/storage/sqlc"
	"Librorum/internal/storage"
	"Librorum/internal/users"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type form struct {
	PublicationStatus string
	Kind              string
	Rating            pgtype.Numeric
	OwnershipStatus   string
	ReadingStatus     string
	CurrentChapter    pgtype.Numeric
	ReadAt            pgtype.Timestamptz
	Notes             string
}

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
	formInput, err := h.formMetadata(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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
		PublicationStatus: formInput.PublicationStatus,
		CoverPath:         coverPath,
		Kind:              formInput.Kind,
		Rating:            formInput.Rating,
		OwnershipStatus:   formInput.OwnershipStatus,
		ReadingStatus:     formInput.ReadingStatus,
		CurrentChapter:    formInput.CurrentChapter,
		ReadAt:            formInput.ReadAt,
		Notes:             formInput.Notes,
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

func (h *Handler) InsertOpenLibraryBooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId, status, err := h.SessionId(ctx, r)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		http.Error(w, "Error parsing multipart form", http.StatusBadRequest)
		return
	}

	formInput, err := h.formMetadata(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	metadata, err := h.formSelectedMetadata(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if metadata.Source == MetadataSourceOpenLibrary && metadata.Description == "" && metadata.WorkKey != "" {
		if description, err := h.Manager.OpenLibrary.WorkDescription(ctx, metadata.WorkKey); err == nil {
			metadata.Description = description
		}
	}

	coverPath := ""
	switch metadata.Source {
	case MetadataSourceCalibre:
		coverPath = h.Manager.safeCachedCoverPath(metadata.CoverPath)
	case MetadataSourceOpenLibrary:
		if metadata.CoverID > 0 {
			coverPath = filepath.Join(h.Manager.CacheDir, fmt.Sprintf("%s_%d.jpg", storage.SanitizeFileName(metadata.Title), metadata.CoverID))
			if cachedPath, err := h.Manager.downloadOpenLibraryCover(ctx, metadata.CoverID, coverPath); err == nil {
				coverPath = cachedPath
			} else {
				h.Logger.Error("Error trying to fetch OpenLibrary cover: "+err.Error(), nil)
				coverPath = ""
			}
		}
	case MetadataSourceGoogleBooks:
		if metadata.CoverURL != "" {
			coverPath = h.Manager.externalCoverPath(metadata.Title, metadata.Source, metadata.SourceID, metadata.CoverURL)
			if cachedPath, err := h.Manager.downloadCoverURL(ctx, metadata.CoverURL, coverPath, "Librorum/0.1"); err == nil {
				coverPath = cachedPath
			} else {
				h.Logger.Error("Error trying to fetch Google Books cover: "+err.Error(), nil)
				coverPath = ""
			}
		}
	}

	response, err := h.Queries.InsertBook(ctx, db.InsertBookParams{
		UserID:            userId,
		Title:             metadata.Title,
		Author:            metadata.Author,
		Description:       metadata.Description,
		Genres:            nonNilStrings(metadata.Genres),
		Language:          metadata.Language,
		PublicationYear:   metadata.PublicationYear,
		PublicationStatus: formInput.PublicationStatus,
		CoverPath:         coverPath,
		Kind:              formInput.Kind,
		Rating:            formInput.Rating,
		OwnershipStatus:   formInput.OwnershipStatus,
		ReadingStatus:     formInput.ReadingStatus,
		CurrentChapter:    formInput.CurrentChapter,
		ReadAt:            formInput.ReadAt,
		Notes:             formInput.Notes,
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

func (h *Handler) SearchOpenLibraryBooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, status, err := h.SessionId(ctx, r)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		http.Error(w, "Error parsing multipart form", http.StatusBadRequest)
		return
	}

	title, author, err := h.formOpenLibrarySearch(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metadata := make([]BookMetadataCandidate, 0, 11)
	calibreMetadata, calibreErr := h.CalibreMetadata(ctx, h.Paths.CoverCacheDir, h.Paths.ImportsDir, title, author)
	if calibreErr != nil {
		h.Logger.Error("Error trying to fetch calibre metadata: "+calibreErr.Error(), nil)
	}

	openLibraryMetadata, openLibraryErr := h.Manager.OpenLibrary.SearchBookMetadataCandidates(ctx, title, author)
	if openLibraryErr != nil {
		h.Logger.Error("Error trying to fetch OpenLibrary metadata: "+openLibraryErr.Error(), nil)
	} else {
		metadata = append(metadata, openLibraryMetadata...)
	}

	googleBooksMetadata, googleBooksErr := h.Manager.GoogleBooks.SearchBookMetadataCandidates(ctx, title, author)
	if googleBooksErr != nil {
		h.Logger.Error("Error trying to fetch Google Books metadata: "+googleBooksErr.Error(), nil)
	} else {
		metadata = append(metadata, googleBooksMetadata...)
	}
	if calibreMetadata != nil && calibreMetadata.CoverPath == "" {
		retryCandidates := calibreRetryCandidates(title, openLibraryMetadata, googleBooksMetadata)
		if len(retryCandidates) > 0 {
			h.Logger.Info("Calibre external ISBN retry candidates", map[string]string{"count": strconv.Itoa(len(retryCandidates)), "candidates": retryCandidateSummary(retryCandidates)})
		}
		updated, err := h.RetryCalibreMetadataCover(ctx, h.Paths.CoverCacheDir, h.Paths.ImportsDir, title, author, calibreMetadata, retryCandidates)
		if err != nil {
			h.Logger.Error("Error trying to retry Calibre metadata with external ISBNs: "+err.Error(), nil)
		} else {
			calibreMetadata = updated
		}
	}
	if calibreMetadata != nil && !titleLikelyMatches(title, calibreMetadata.Title) {
		h.Logger.Info("Rejected Calibre metadata for mismatched title", map[string]string{"requested_title": title, "returned_title": calibreMetadata.Title})
		if calibreMetadata.CoverPath != "" {
			_ = os.Remove(calibreMetadata.CoverPath)
		}
		calibreMetadata = nil
	}
	if calibreMetadata != nil {
		metadata = append([]BookMetadataCandidate{*calibreMetadata}, metadata...)
	}

	if len(metadata) == 0 {
		if calibreErr != nil || openLibraryErr != nil || googleBooksErr != nil {
			http.Error(w, "Error trying to fetch metadata", http.StatusBadGateway)
			return
		}
		http.Error(w, "No metadata found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(metadata); err != nil {
		h.Logger.Error("Error trying to write metadata search response: "+err.Error(), nil)
	}
}

func (h *Handler) formMetadata(r *http.Request) (*form, error) {
	kind := r.FormValue("kind")
	if kind == "" {
		kind = "book"
	}
	ratingForm := strings.TrimSpace(r.FormValue("rating"))

	var rating pgtype.Numeric
	if ratingForm == "" {
		rating = pgtype.Numeric{Valid: false}
	} else if err := rating.Scan(ratingForm); err != nil {
		return nil, errors.New("invalid rating")
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
		return nil, errors.New("invalid current chapter")
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
			return nil, errors.New("invalid read at value")
		}
		readAt = pgtype.Timestamptz{Time: parsed, Valid: true}
	}

	notes := r.FormValue("notes")

	formData := &form{
		PublicationStatus: publication_status,
		Kind:              kind,
		Rating:            rating,
		OwnershipStatus:   ownership_status,
		ReadingStatus:     reading_status,
		CurrentChapter:    currentChapter,
		ReadAt:            readAt,
		Notes:             notes,
	}
	return formData, nil
}

func (h *Handler) formOpenLibrarySearch(r *http.Request) (string, string, error) {
	title := strings.TrimSpace(r.FormValue("title"))
	author := strings.TrimSpace(r.FormValue("author"))
	if title == "" {
		return "", "", errors.New("title is required")
	}
	return title, author, nil
}

func calibreRetryCandidates(queryTitle string, groups ...[]BookMetadataCandidate) []calibreRetryCandidate {
	candidates := make([]calibreRetryCandidate, 0)
	for _, group := range groups {
		for _, candidate := range group {
			if !titleLikelyMatches(queryTitle, candidate.Title) {
				continue
			}
			if candidate.ISBN == "" && len(candidate.ISBNs) == 0 {
				candidates = append(candidates, calibreRetryCandidate{Title: candidate.Title})
				continue
			}
			for _, isbn := range append([]string{candidate.ISBN}, candidate.ISBNs...) {
				candidates = append(candidates, calibreRetryCandidate{Title: candidate.Title, ISBN: isbn})
			}
		}
	}
	return uniqueCalibreRetryCandidates(candidates)
}

func retryCandidateSummary(candidates []calibreRetryCandidate) string {
	parts := make([]string, 0, len(candidates))
	for i, candidate := range candidates {
		if i >= 12 {
			parts = append(parts, "...")
			break
		}
		if candidate.ISBN == "" {
			parts = append(parts, candidate.Title)
			continue
		}
		parts = append(parts, candidate.Title+":"+candidate.ISBN)
	}
	return strings.Join(parts, ", ")
}

func (h *Handler) formSelectedMetadata(r *http.Request) (*BookMetadataCandidate, error) {
	title := strings.TrimSpace(r.FormValue("selected_title"))
	if title == "" {
		return nil, errors.New("selected title is required")
	}

	source := strings.TrimSpace(r.FormValue("selected_source"))
	switch source {
	case MetadataSourceCalibre, MetadataSourceOpenLibrary, MetadataSourceGoogleBooks:
	case "":
		return nil, errors.New("selected metadata source is required")
	default:
		return nil, errors.New("invalid selected metadata source")
	}

	metadata := &BookMetadataCandidate{
		Source:      source,
		SourceID:    strings.TrimSpace(r.FormValue("selected_source_id")),
		Title:       title,
		Author:      strings.TrimSpace(r.FormValue("selected_author")),
		Description: strings.TrimSpace(r.FormValue("selected_description")),
		Genres:      nonNilStrings(NormalizeGenres(r.Form["selected_genres"])),
		Language:    strings.TrimSpace(r.FormValue("selected_language")),
		CoverURL:    strings.TrimSpace(r.FormValue("selected_cover_url")),
		CoverPath:   strings.TrimSpace(r.FormValue("selected_cover_path")),
		WorkKey:     strings.TrimSpace(r.FormValue("selected_work_key")),
	}
	if metadata.SourceID == "" && metadata.WorkKey != "" {
		metadata.SourceID = metadata.WorkKey
	}

	if coverID := strings.TrimSpace(r.FormValue("selected_cover_id")); coverID != "" {
		parsed, err := strconv.Atoi(coverID)
		if err != nil {
			return nil, errors.New("invalid selected cover id")
		}
		metadata.CoverID = parsed
	}
	if year := strings.TrimSpace(r.FormValue("selected_publication_year")); year != "" {
		parsed, err := strconv.ParseInt(year, 10, 32)
		if err != nil {
			return nil, errors.New("invalid selected publication year")
		}
		publicationYear := int32(parsed)
		metadata.PublicationYear = &publicationYear
	}

	return metadata, nil
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
