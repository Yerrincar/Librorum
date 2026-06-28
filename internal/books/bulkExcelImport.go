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
	"strings"
)

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

		_, err = h.Queries.InsertBook(ctx, db.InsertBookParams{
			UserID:            userId,
			Kind:              "book",
			Title:             bookMetadata.Title,
			Author:            bookMetadata.Author,
			Description:       bookMetadata.Description,
			Genres:            nonNilStrings(bookMetadata.Genres),
			Language:          bookMetadata.Language,
			PublicationYear:   bookMetadata.PublicationYear,
			CoverPath:         bookMetadata.CoverPath,
			OwnershipStatus:   "none",
			ReadingStatus:     "unread",
			PublicationStatus: "unknown",
		})
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
