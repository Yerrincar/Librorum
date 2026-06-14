package books

import (
	"Librorum/internal/storage"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (h *Handler) FileUploader(r *http.Request) (string, error) {

	file, handler, err := r.FormFile("file")
	if err != nil {
		h.Logger.Error("Error trying to retrieve file"+err.Error(), nil)
		return "", err
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		h.Logger.Error("Invalid file"+err.Error(), nil)
		return "", err
	}

	if !isValidFileType(fileBytes, handler.Filename) {
		h.Logger.Error("Invalid file", nil)
		return "", fmt.Errorf("invalid file type: %s", handler.Filename)
	}

	dst, err := h.createFile(handler.Filename)
	if err != nil {
		h.Logger.Error("Error trying to downlaod file"+err.Error(), nil)
		return "", err
	}

	defer dst.Close()

	if _, err := dst.Write(fileBytes); err != nil {
		h.Logger.Error("Error trying to save the file"+err.Error(), nil)
		return "", err
	}

	h.Logger.Info("Uploaded File: "+handler.Filename+"\n", nil)
	h.Logger.Info("File Size: "+strconv.FormatInt(handler.Size, 10)+"\n", nil)
	return filepath.Join(storage.NewPaths(h.Paths.RootDir).BooksDir, storage.SanitizeFileName(handler.Filename)), nil
}

func (h *Handler) createFile(filename string) (*os.File, error) {
	dst, err := os.Create(filepath.Join(storage.NewPaths(h.Paths.RootDir).BooksDir, storage.SanitizeFileName(filename)))
	if err != nil {
		return nil, err
	}
	return dst, nil
}

func isValidFileType(file []byte, filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".epub" {
		return false
	}

	fileType := http.DetectContentType(file)
	return fileType == "application/epub+zip" ||
		fileType == "application/zip" ||
		fileType == "application/octet-stream"
}
