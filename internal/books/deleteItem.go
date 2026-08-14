package books

import (
	db "Librorum/internal/platform/storage/sqlc"
	"encoding/json"
	"net/http"
	"strconv"
)

func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
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

	response, err := h.Queries.DeleteItemFromLibrary(ctx, db.DeleteItemFromLibraryParams{ID: bookID, UserID: userId})
	if err != nil {
		h.Logger.Error("Error trying to delete book in the database: "+err.Error(), nil)
		http.Error(w, "Error trying to delete book in the database", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}
