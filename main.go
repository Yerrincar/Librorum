package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"Librorum/internal/storage"
)

type Config struct {
	Addr      string
	DataDir   string
	OLContact string
}

type App struct {
	config Config
	paths  storage.Paths
}

func main() {
	config := configFromEnv()
	paths := storage.NewPaths(config.DataDir)
	if err := storage.EnsureDirs(paths); err != nil {
		log.Fatalf("prepare storage directories: %v", err)
	}

	app := &App{
		config: config,
		paths:  paths,
	}

	server := &http.Server{
		Addr:              config.Addr,
		Handler:           app.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("librorum listening on %s", config.Addr)
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
	mux.HandleFunc("GET /api/example/library-items", a.handleExampleLibraryItems)
	return logRequests(mux)
}

func (a *App) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) handleExampleLibraryItems(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"message": "replace this with your sqlc-backed library item handler",
		"storage": map[string]string{
			"books_dir":  a.paths.BooksDir,
			"covers_dir": a.paths.CoverCacheDir,
		},
	})
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
