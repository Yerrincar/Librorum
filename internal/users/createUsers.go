package users

import (
	"Librorum/internal/helpers"
	db "Librorum/internal/platform/storage/sqlc"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func (u *UserHandle) Register(w http.ResponseWriter, r *http.Request) {
	appCtx := r.Context()
	setupCtx, cancel := context.WithTimeout(appCtx, 10*time.Second)
	defer cancel()
	json.NewDecoder(r.Body).Decode(&u.User)
	err := u.User.Password.Set(*u.User.Password.plaintext)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	v := helpers.New()
	if ValidateUser(v, u.User); !v.Valid() {
		http.Error(w, "Error validating user", http.StatusBadRequest)
		return
	}

	_, err = u.Queries.InsertUser(setupCtx, db.InsertUserParams{
		Username:     u.User.Username,
		Email:        u.User.Email,
		PasswordHash: u.User.Password.hash,
		DisplayName:  u.User.DisplayName,
	})
	if err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
		http.Error(w, "Email already in use", http.StatusBadRequest)
		log.Print(err.Error())
	}
	if err != nil {
		http.Error(w, "Error trying to register user", http.StatusBadRequest)
		log.Print(err.Error())
	}
}
