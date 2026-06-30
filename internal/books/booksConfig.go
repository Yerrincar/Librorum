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
	ID                 int64      `json:"id"`
	Title              string     `json:"Title"`
	Author             string     `json:"Author"`
	Kind               string     `json:"Kind"`
	Description        string     `json:"Description"`
	Language           string     `json:"Language"`
	Publication_year   string     `json:"Publication_year"`
	Genres             []string   `json:"Genres"`
	Rating             *float64   `json:"Rating,omitempty"`
	Ownership_status   string     `json:"Ownership_status"`
	Reading_status     string     `json:"Reading_status"`
	Publication_status string     `json:"Publication_status"`
	Current_chapter    *float64   `json:"Current_chapter,omitempty"`
	Total_chapters     *float64   `json:"Total_chapters,omitempty"`
	Read_at            *time.Time `json:"Read_at,omitempty"`
	Finished_at        *time.Time `json:"Finished_at,omitempty"`
	Cover_path         string     `json:"Cover_path"`
	Notes              string     `json:"Notes"`
}
type Logger interface {
	Info(message string, properties map[string]string)
	Error(message string, properties map[string]string)
	Debug(message string, properties map[string]string)
}
