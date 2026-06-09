package users

import (
	db "Librorum/internal/platform/storage/sqlc"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Logger interface {
	Info(message string, properties map[string]string)
	Error(message string, properties map[string]string)
	Debug(message string, properties map[string]string)
}

type SessionConfig struct {
	HMC               string
	SecretKey         []byte
	SessionExpiration time.Duration
}

type UserProfile struct {
	Id          int       `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Password    password  `json:"-"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type password struct {
	Plaintext *string
	Hash      string
}

type UserHandle struct {
	DB            *pgxpool.Pool
	Queries       *db.Queries
	User          *UserProfile
	Logger        Logger
	SessionConfig SessionConfig
}

var (
	ErrDuplicatedEmail = errors.New("That email is already in use")
)
