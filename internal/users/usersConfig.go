package users

import (
	db "Librorum/internal/platform/storage/sqlc"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserProfile struct {
	Id          int       `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Password    password  `json:"-"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type password struct {
	plaintext *string
	hash      string
}

type UserHandle struct {
	DB      *pgxpool.Pool
	Queries *db.Queries
	User    *UserProfile
}

var (
	ErrDuplicatedEmail = errors.New("That email is already in use")
)
