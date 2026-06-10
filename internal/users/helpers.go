package users

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

const sessionCookieName = "sessionId"

func GenerateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func HashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func SessionTokenFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", err
	}
	if cookie.Value == "" {
		return "", errors.New("empty session cookie")
	}
	return cookie.Value, nil
}

func (u *UserHandle) GetUserByUserID(ctx context.Context, id int64) (*UserProfile, error) {
	dbUser, err := u.Queries.SelectUserByID(ctx, id)
	if err == pgx.ErrNoRows {
		u.Logger.Error("The username was not found: "+err.Error(), map[string]string{
			"id": strconv.FormatInt(id, 10),
		})
		return nil, err
	}
	if err != nil {
		u.Logger.Error("Error trying to select username: "+err.Error(), map[string]string{
			"id": strconv.FormatInt(id, 10),
		})
		return nil, err
	}
	user := &UserProfile{
		Id:          int(dbUser.ID),
		Username:    dbUser.Username,
		Email:       dbUser.Email,
		DisplayName: dbUser.DisplayName,
	}
	return user, nil

}

func (u *UserHandle) GetUserByUsername(username string, ctx context.Context) (*UserProfile, error) {
	result, err := u.Queries.SelectUserByUsername(ctx, username)
	if err == pgx.ErrNoRows {
		u.Logger.Error("The username was not found: "+err.Error(), map[string]string{
			"username": username,
		})
		return nil, err
	}
	if err != nil {
		u.Logger.Error("Error trying to select username: "+err.Error(), map[string]string{
			"username": username,
		})
		return nil, err
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
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json response: %v", err)
	}
}
