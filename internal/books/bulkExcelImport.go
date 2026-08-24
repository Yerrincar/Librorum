package books

import (
	db "Librorum/internal/platform/storage/sqlc"
	"Librorum/internal/storage"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var excelKindAliases = map[string]string{
	"book":   "book",
	"books":  "book",
	"libro":  "book",
	"manga":  "manga",
	"manhwa": "manhwa",
}

var excelOwnershipStatusAliases = map[string]string{
	"none":                       "none",
	"not_owned":                  "none",
	"unowned":                    "none",
	"owned":                      "owned_physical",
	"owned_physical":             "owned_physical",
	"physical":                   "owned_physical",
	"paper":                      "owned_physical",
	"paperback":                  "owned_physical",
	"hardcover":                  "owned_physical",
	"owned_digital":              "owned_digital",
	"digital":                    "owned_digital",
	"ebook":                      "owned_digital",
	"e_book":                     "owned_digital",
	"epub":                       "owned_digital",
	"pdf":                        "owned_digital",
	"kindle":                     "owned_digital",
	"owned_physical_and_digital": "owned_physical_and_digital",
	"physical_and_digital":       "owned_physical_and_digital",
	"digital_and_physical":       "owned_physical_and_digital",
	"both":                       "owned_physical_and_digital",
	"wishlist":                   "wishlist",
	"wish_list":                  "wishlist",
	"to_buy":                     "wishlist",
}

var excelReadingStatusAliases = map[string]string{
	"unread":            "unread",
	"not_read":          "unread",
	"pending":           "to_read",
	"to_read":           "to_read",
	"reading":           "reading",
	"currently_reading": "reading",
	"in_progress":       "reading",
	"read":              "read",
	"finished":          "read",
	"completed":         "read",
	"done":              "read",
	"dropped":           "dropped",
	"dnf":               "dropped",
}

var excelPublicationStatusAliases = map[string]string{
	"unknown":     "unknown",
	"finished":    "finished",
	"complete":    "finished",
	"completed":   "finished",
	"ongoing":     "ongoing",
	"in_progress": "ongoing",
	"hiatus":      "hiatus",
	"on_hiatus":   "hiatus",
}

type BulkExcelImportResponse struct {
	ImportedCount int      `json:"imported_count"`
	SkippedCount  int      `json:"skipped_count"`
	Imported      []string `json:"imported"`
	Skipped       []string `json:"skipped"`
}

func (h *Handler) BulkExcelImport(w http.ResponseWriter, r *http.Request) {
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

	excelPath, err := h.uploadExcelFile(r)
	if err != nil {
		http.Error(w, "Error trying to upload Excel file", http.StatusBadRequest)
		return
	}

	spreadsheet, err := h.formSpreadSheet(r)
	if err != nil {
		http.Error(w, "Error trying to form spreadsheet", http.StatusBadRequest)
		return
	}

	excelData, err := ExcelTitleAuthor(excelPath, spreadsheet)
	if err != nil {
		h.Logger.Error("Error trying to read Excel data: "+err.Error(), nil)
		http.Error(w, "Error trying to read Excel data", http.StatusBadRequest)
		return
	}

	response := BulkExcelImportResponse{
		Imported: make([]string, 0, len(excelData)),
		Skipped:  make([]string, 0),
	}
	for _, row := range excelData {
		bookMetadata, err := h.CalibreMetadata(ctx, h.Paths.CoverCacheDir, h.Paths.ImportsDir, row.Title, row.Author)
		if err != nil {
			h.Logger.Error("Error trying to fetch Calibre metadata: "+err.Error(), map[string]string{"title": row.Title, "author": row.Author})
			response.SkippedCount++
			response.Skipped = append(response.Skipped, row.Title)
			continue
		}
		if bookMetadata == nil || strings.TrimSpace(bookMetadata.Title) == "" {
			response.SkippedCount++
			response.Skipped = append(response.Skipped, row.Title)
			continue
		}

		_, err = h.Queries.InsertBook(ctx, h.excelInsertBookParams(userId, row, bookMetadata))
		if err != nil {
			h.Logger.Error("Error trying to insert book in the database: "+err.Error(), map[string]string{"title": bookMetadata.Title})
			response.SkippedCount++
			response.Skipped = append(response.Skipped, row.Title)
			continue
		}
		response.ImportedCount++
		response.Imported = append(response.Imported, bookMetadata.Title)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) excelInsertBookParams(userID int64, row ExcelBookRow, metadata *BookMetadataCandidate) db.InsertBookParams {
	rating, err := parseOptionalNumeric(row.Rating, "rating")
	if err != nil {
		h.Logger.Debug("Ignoring invalid Excel rating", map[string]string{"title": row.Title, "rating": row.Rating})
	}
	currentChapter, err := parseOptionalNumeric(row.CurrentChapter, "current chapter")
	if err != nil {
		h.Logger.Debug("Ignoring invalid Excel current chapter", map[string]string{"title": row.Title, "current_chapter": row.CurrentChapter})
	}
	totalChapters, err := parseOptionalNumeric(row.TotalChapters, "total chapters")
	if err != nil {
		h.Logger.Debug("Ignoring invalid Excel total chapters", map[string]string{"title": row.Title, "total_chapters": row.TotalChapters})
	}
	readAt, err := parseOptionalTimestamp(row.ReadAt)
	if err != nil {
		h.Logger.Debug("Ignoring invalid Excel read_at", map[string]string{"title": row.Title, "read_at": row.ReadAt})
	}

	publicationYear := metadata.PublicationYear
	if excelYear := parseExcelPublicationYear(row.PublicationYear); excelYear != nil {
		publicationYear = excelYear
	}

	ownershipValue := row.OwnershipStatus
	if strings.TrimSpace(ownershipValue) == "" {
		ownershipValue = row.Format
	}

	return db.InsertBookParams{
		UserID:            userID,
		Kind:              excelEnum(row.Kind, "book", excelKindAliases),
		Title:             defaultString(metadata.Title, row.Title),
		Author:            defaultString(metadata.Author, row.Author),
		Description:       metadata.Description,
		Genres:            nonNilStrings(metadata.Genres),
		Language:          metadata.Language,
		PublicationYear:   publicationYear,
		CoverPath:         metadata.CoverPath,
		Rating:            rating,
		OwnershipStatus:   excelEnum(ownershipValue, "none", excelOwnershipStatusAliases),
		ReadingStatus:     excelEnum(row.ReadingStatus, "unread", excelReadingStatusAliases),
		PublicationStatus: excelEnum(row.PublicationStatus, "unknown", excelPublicationStatusAliases),
		CurrentChapter:    currentChapter,
		TotalChapters:     totalChapters,
		ReadAt:            readAt,
		Notes:             excelNotes(row),
	}
}

func excelEnum(value, fallback string, aliases map[string]string) string {
	value = normalizeExcelChoice(value)
	if value == "" {
		return fallback
	}
	if canonical, ok := aliases[value]; ok {
		return canonical
	}
	return fallback
}

func normalizeExcelChoice(value string) string {
	value = strings.NewReplacer("&", " and ", "+", " and ", "(", " ", ")", " ", ",", " ").Replace(value)
	return normalizeExcelHeader(value)
}

func parseExcelPublicationYear(value string) *int32 {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if year, ok := validExcelYear(value); ok {
		return &year
	}
	if len(value) >= 4 {
		if year, ok := validExcelYear(value[:4]); ok {
			return &year
		}
	}
	if parsed, err := strconv.ParseFloat(value, 64); err == nil {
		year := int64(parsed)
		if parsed == float64(year) {
			return int64ToExcelYear(year)
		}
	}
	return nil
}

func validExcelYear(value string) (int32, bool) {
	year, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, false
	}
	parsed := int64ToExcelYear(year)
	if parsed == nil {
		return 0, false
	}
	return *parsed, true
}

func int64ToExcelYear(year int64) *int32 {
	if year < 0 || year > 3000 {
		return nil
	}
	parsed := int32(year)
	return &parsed
}

func excelNotes(row ExcelBookRow) string {
	notes := strings.TrimSpace(row.Notes)
	format := strings.TrimSpace(row.Format)
	if format == "" {
		return notes
	}
	if notes == "" {
		return "Format: " + format
	}
	return notes + "\nFormat: " + format
}

func (h *Handler) formSpreadSheet(r *http.Request) (string, error) {
	spreadsheet := r.FormValue("spreadsheet")
	if spreadsheet == "" {
		spreadsheet = "Inventario"
	}
	return spreadsheet, nil
}

func (h *Handler) uploadExcelFile(r *http.Request) (string, error) {
	file, handler, err := r.FormFile("file")
	if err != nil {
		h.Logger.Error("Error trying to retrieve file"+err.Error(), nil)
		return "", err
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(handler.Filename))
	if ext != ".xlsx" && ext != ".xlsm" {
		return "", fmt.Errorf("invalid Excel file type: %s", handler.Filename)
	}

	dstPath, _, err := storage.UniquePath(h.Paths.ImportsDir, handler.Filename)
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
	h.Logger.Info("Uploaded Excel File: "+handler.Filename, nil)
	return dstPath, nil
}
