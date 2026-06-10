package users

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func (u *UserHandle) CurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionToken, err := SessionTokenFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized status", http.StatusUnauthorized)
		return
	}

	session, err := u.Queries.FindSessionByTokenHash(ctx, HashSessionToken(sessionToken))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Unauthorized status", http.StatusUnauthorized)
			return
		}
		u.Logger.Error("Error trying to find session: "+err.Error(), nil)
		http.Error(w, "There was a problem and we couldn't fulfill your request", http.StatusInternalServerError)
		return
	}
	dbUser, err := u.GetUserByUserID(ctx, session.UserID)
	if err != nil {
		u.Logger.Error("Error trying to get user: "+err.Error(), nil)
		http.Error(w, "There was a problem and we couldn't fulfill your request", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, http.StatusOK, dbUser)
	u.Logger.Info("User retrieved successfully", map[string]string{
		"username": dbUser.Username,
	})
}
