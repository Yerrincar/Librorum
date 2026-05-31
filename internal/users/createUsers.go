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

	var input struct {
		Username    string `json:"username"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
	}
	json.NewDecoder(r.Body).Decode(&input)
	user := &UserProfile{
		Username:    input.Username,
		Email:       input.Email,
		DisplayName: input.DisplayName,
	}
	err := user.Password.Set(input.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		log.Printf("Error hashing password: %v", err)
		return
	}
	v := helpers.New()
	if ValidateUser(v, user); !v.Valid() {
		http.Error(w, "Error validating user", http.StatusBadRequest)
		log.Printf("Error validating user: %v", v.Errors)
		return
	}

	response, err := u.Queries.InsertUser(setupCtx, db.InsertUserParams{
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.Password.Hash,
		DisplayName:  user.DisplayName,
	})
	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
			http.Error(w, "Email already in use", http.StatusBadRequest)
			log.Print(err.Error())
		}
		http.Error(w, "Error trying to register user", http.StatusBadRequest)
		log.Print(err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}
