package books

import (
	db "Librorum/internal/platform/storage/sqlc"
	"Librorum/internal/storage"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	Db      *pgxpool.Pool
	Queries *db.Queries
	Book    Book
	Logger  Logger
	Manager *Manager
	Paths   storage.Paths
}

type Book struct {
	Title              string    `json:"Title"`
	Author             string    `json:"Author"`
	Kind               string    `json:"Kind"`
	Description        string    `json:"Description"`
	Language           string    `json:"Language"`
	Publication_year   string    `json:"Publication_year"`
	Genres             []string  `json:"Genres"`
	Rating             float64   `json:"Rating"`
	Ownership_status   string    `json:"Ownership_status"`
	Reading_status     string    `json:"Reading_status"`
	Publication_status string    `json:"Publication_status"`
	Current_chapter    int       `json:"Current_chapter"`
	Total_chapters     int       `json:"Total_chapters"`
	Read_at            time.Time `json:"Read_at"`
	Finished_at        time.Time `json:"Finished_at"`
	Cover_path         string    `json:"Cover_path"`
	Notes              string    `json:"Notes"`
}
type Logger interface {
	Info(message string, properties map[string]string)
	Error(message string, properties map[string]string)
	Debug(message string, properties map[string]string)
}
