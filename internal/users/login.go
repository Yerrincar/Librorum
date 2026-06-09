package users

import (
	db "Librorum/internal/platform/storage/sqlc"
	"bytes"
	"context"
	"database/sql"
	"encoding/gob"
	"encoding/json"
	"log"
	"net/http"
)

func (u *UserHandle) LoginUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	json.NewDecoder(r.Body).Decode(&input)

	dbUser, err := u.GetUsername(input.Username, ctx)
	if err != nil {
		u.Logger.Error("Error trying to get the user's username: "+err.Error(), map[string]string{
			"username": input.Username,
		})
		return
	}

	match, err := dbUser.Password.Matches(input.Password)
	if err != nil {
		u.Logger.Error("Error trying to match passwords: "+err.Error(), map[string]string{
			"username": input.Username,
		})
		return
	}

	if !match {
		u.Logger.Info("Password and username combination doesn't exist", map[string]string{
			"username": input.Username,
		})
		return
	}

	userId := UserProfile{
		Id: dbUser.Id,
	}

	var buf bytes.Buffer

	err = gob.NewEncoder(&buf).Encode(&userId)
	if err != nil {
		u.Logger.Error("Error trying to encode data: "+err.Error(), nil)
	}

	session := buf.String()

	_, err = u.Queries.CreateSession(ctx, db.CreateSessionParams{
		UserID: int64(userId.Id),
		Hash:   &session,
	})
	if err != nil {
		u.Logger.Error("Error trying to create session: "+err.Error(), nil)
	}

	cookie := http.Cookie{
		Name:     "sessionId",
		Value:    session,
		Path:     "/",
		MaxAge:   int(u.SessionConfig.SessionExpiration.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	err = WriteEncrypted(w, cookie, u.SessionConfig.SecretKey)
	if err != nil {
		u.Logger.Error("Error trying to set up cookie: "+err.Error(), nil)
	}
	writeJSON(w, http.StatusOK, dbUser)
	u.Logger.Info("Logged sucessfuly", map[string]string{
		"username": dbUser.Username,
	})
}
func (u *UserHandle) GetUsername(username string, ctx context.Context) (*UserProfile, error) {
	result, err := u.Queries.SelectUserByUsername(ctx, username)
	if err == sql.ErrNoRows {
		u.Logger.Error("The username was not found: "+err.Error(), map[string]string{
			"username": username,
		})
		return nil, err
	}
	if err != nil {
		u.Logger.Error("Error trying to select username: "+err.Error(), map[string]string{
			"username": username,
		})
	}
	pass := &password{
		Hash: result.PasswordHash,
	}
	user := &UserProfile{
		Id:       int(result.ID),
		Username: result.Username,
		Password: *pass,
	}
	return user, nil
}
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json response: %v", err)
	}
}
