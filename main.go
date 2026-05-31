package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"Librorum/internal/books"
	db "Librorum/internal/platform/storage/sqlc"
	"Librorum/internal/storage"
	"Librorum/internal/users"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Addr      string
	DataDir   string
	OLContact string
}

type App struct {
	config      Config
	paths       storage.Paths
	bookHandler *books.Handler
	userHandler *users.UserHandle
}

func main() {
	appCtx := context.Background()
	setupCtx, cancel := context.WithTimeout(appCtx, 10*time.Second)
	defer cancel()

	databaseURL := envOrDefault("LIBRORUM_DATABASE_URL", "")
	if databaseURL == "" {
		log.Fatal("LIBRORUM_DATABASE_URL is required")
	}

	dbPool, err := pgxpool.New(setupCtx, databaseURL)
	if err != nil || dbPool.Ping(setupCtx) != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}

	defer dbPool.Close()
	config := configFromEnv()
	paths := storage.NewPaths(config.DataDir)
	if err := storage.EnsureDirs(paths); err != nil {
		log.Fatalf("prepare storage directories: %v", err)
	}

	h := &books.Handler{
		Db:      dbPool,
		Queries: db.New(dbPool),
	}
	u := &users.UserHandle{
		DB:      dbPool,
		Queries: db.New(dbPool),
	}

	app := &App{
		config:      config,
		paths:       paths,
		bookHandler: h,
		userHandler: u,
	}

	server := &http.Server{
		Addr:              config.Addr,
		Handler:           app.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("librorum listening on %s", config.Addr)
		log.Printf("database url configured: %t", os.Getenv("LIBRORUM_DATABASE_URL") != "")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Print("librorum shutting down")
}

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", a.handleHealth)
	mux.HandleFunc("GET /api/example/library-items", a.bookHandler.DisplayBooks)
	mux.HandleFunc("POST /users/register", a.userHandler.Register)
	return logRequests(mux)
}

func (a *App) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func configFromEnv() Config {
	return Config{
		Addr:      envOrDefault("LIBRORUM_ADDR", ":8080"),
		DataDir:   envOrDefault("LIBRORUM_DATA_DIR", "data"),
		OLContact: os.Getenv("LIBRORUM_OPENLIBRARY_CONTACT"),
	}
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json response: %v", err)
	}
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
