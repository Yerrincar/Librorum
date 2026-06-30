package books

import (
	db "Librorum/internal/platform/storage/sqlc"
	"Librorum/internal/storage"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) UpdateFields(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId, status, err := h.SessionId(ctx, r)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	bookID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || bookID < 1 {
		http.Error(w, "invalid book id", http.StatusBadRequest)
		return
	}

	formInput, err := h.formUpdateLibraryItem(r, bookID, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	coverPath := formInput.CoverPath
	uploadedCoverPath, err := h.uploadCoverFile(r)
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if uploadedCoverPath != "" {
		coverPath = uploadedCoverPath
	}

	response, err := h.Queries.UpdateLibraryItems(ctx, db.UpdateLibraryItemsParams{
		ID:              bookID,
		UserID:          userId,
		Title:           formInput.Title,
		Author:          formInput.Author,
		Description:     formInput.Description,
		Genres:          nonNilStrings(formInput.Genres),
		Language:        formInput.Language,
		CoverPath:       coverPath,
		Rating:          formInput.Rating,
		OwnershipStatus: formInput.OwnershipStatus,
		ReadingStatus:   formInput.ReadingStatus,
		CurrentChapter:  formInput.CurrentChapter,
		TotalChapters:   formInput.TotalChapters,
		ReadAt:          formInput.ReadAt,
		Notes:           formInput.Notes,
	})
	if err != nil {
		h.Logger.Error("Error trying to update book in the database: "+err.Error(), nil)
		http.Error(w, "Error trying to update book in the database", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) uploadCoverFile(r *http.Request) (string, error) {
	file, handler, err := r.FormFile("file")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return "", err
		}
		h.Logger.Error("Error trying to retrieve file"+err.Error(), nil)
		return "", err
	}
	defer file.Close()

	filename := storage.SanitizeFileName(handler.Filename)

	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", fmt.Errorf("invalid cover file type: %s", filename)
	}
	if err := validateCoverImage(file, ext); err != nil {
		return "", err
	}

	dstPath, _, err := storage.UniquePath(h.Paths.CoverCacheDir, filename)
	if err != nil {
		return "", err
	}
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}
	if err := dst.Sync(); err != nil {
		return "", err
	}
	h.Logger.Info("Uploaded cover File: "+filename, nil)
	return dstPath, nil
}

func validateCoverImage(file interface {
	io.Reader
	io.Seeker
}, ext string) error {
	var header [512]byte
	n, err := file.Read(header[:])
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	contentType := http.DetectContentType(header[:n])
	if contentType == "image/jpeg" && (ext == ".jpg" || ext == ".jpeg") {
		return nil
	}
	if contentType == "image/png" && ext == ".png" {
		return nil
	}
	return fmt.Errorf("invalid cover file content type: %s", contentType)
}

func (h *Handler) formUpdateLibraryItem(r *http.Request, id, userID int64) (db.UpdateLibraryItemsParams, error) {
	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		return db.UpdateLibraryItemsParams{}, errors.New("title is required")
	}

	rating, err := parseOptionalNumeric(r.FormValue("rating"), "rating")
	if err != nil {
		return db.UpdateLibraryItemsParams{}, err
	}

	currentChapter, err := parseOptionalNumeric(r.FormValue("current_chapter"), "current chapter")
	if err != nil {
		return db.UpdateLibraryItemsParams{}, err
	}
	totalChapters, err := parseOptionalNumeric(r.FormValue("total_chapters"), "total chapters")
	if err != nil {
		return db.UpdateLibraryItemsParams{}, err
	}

	readAt, err := parseOptionalTimestamp(r.FormValue("read_at"))
	if err != nil {
		h.Logger.Error("Invalid read_at value: "+err.Error(), map[string]string{"read_at": r.FormValue("read_at")})
		return db.UpdateLibraryItemsParams{}, errors.New("invalid read at value")
	}

	return db.UpdateLibraryItemsParams{
		ID:              id,
		UserID:          userID,
		Title:           title,
		Author:          strings.TrimSpace(r.FormValue("author")),
		Rating:          rating,
		CoverPath:       strings.TrimSpace(r.FormValue("cover_path")),
		ReadAt:          readAt,
		Description:     strings.TrimSpace(r.FormValue("description")),
		Language:        strings.TrimSpace(r.FormValue("language")),
		Genres:          parseGenres(r),
		OwnershipStatus: defaultString(r.FormValue("ownership_status"), "none"),
		ReadingStatus:   defaultString(r.FormValue("reading_status"), "unread"),
		CurrentChapter:  currentChapter,
		TotalChapters:   totalChapters,
		Notes:           strings.TrimSpace(r.FormValue("notes")),
	}, nil
}

func parseOptionalNumeric(value string, field string) (pgtype.Numeric, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return pgtype.Numeric{Valid: false}, nil
	}

	var numeric pgtype.Numeric
	if err := numeric.Scan(value); err != nil {
		return pgtype.Numeric{}, fmt.Errorf("invalid %s", field)
	}
	return numeric, nil
}

func parseOptionalTimestamp(value string) (pgtype.Timestamptz, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return pgtype.Timestamptz{Valid: false}, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		parsed, err = time.ParseInLocation("2006-01-02T15:04", value, time.Local)
	}
	if err != nil {
		return pgtype.Timestamptz{}, err
	}

	return pgtype.Timestamptz{Time: parsed, Valid: true}, nil
}

func parseGenres(r *http.Request) []string {
	values := r.Form["genres"]
	if len(values) == 0 {
		values = strings.Split(r.FormValue("genres"), ",")
	}

	genres := make([]string, 0, len(values))
	for _, value := range values {
		for _, genre := range strings.Split(value, ",") {
			genre = strings.TrimSpace(genre)
			if genre != "" {
				genres = append(genres, genre)
			}
		}
	}
	return genres
}

func defaultString(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
