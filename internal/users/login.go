package users

import (
	db "Librorum/internal/platform/storage/sqlc"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func (u *UserHandle) LoginUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	dbUser, err := u.GetUserByUsername(input.Username, ctx)
	if err != nil {
		u.Logger.Error("Error trying to get the user's username: "+err.Error(), map[string]string{
			"username": input.Username,
		})
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	match, err := dbUser.Password.Matches(input.Password)
	if err != nil {
		u.Logger.Error("Error trying to match passwords: "+err.Error(), map[string]string{
			"username": input.Username,
		})
		http.Error(w, "There was a problem logging you in", http.StatusInternalServerError)
		return
	}

	if !match {
		u.Logger.Info("Password and username combination doesn't exist", map[string]string{
			"username": input.Username,
		})
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	sessionToken, err := GenerateSessionToken()
	if err != nil {
		u.Logger.Error("Error trying to generate session token: "+err.Error(), nil)
		http.Error(w, "There was a problem logging you in", http.StatusInternalServerError)
		return
	}
	tokenHash := HashSessionToken(sessionToken)

	_, err = u.Queries.CreateSession(ctx, db.CreateSessionParams{
		UserID:    int64(dbUser.Id),
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(u.SessionConfig.SessionExpiration),
			Valid: true,
		},
	})
	if err != nil {
		u.Logger.Error("Error trying to create session: "+err.Error(), nil)
		http.Error(w, "There was a problem logging you in", http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionToken,
		Path:     "/",
		MaxAge:   int(u.SessionConfig.SessionExpiration.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)
	WriteJSON(w, http.StatusOK, dbUser)
	u.Logger.Info("Logged sucessfuly", map[string]string{
		"username": dbUser.Username,
	})
}
