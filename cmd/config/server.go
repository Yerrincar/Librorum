package config

import (
	"Librorum/internal/books"
	users "Librorum/internal/users/register"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	BookHandler *books.Handler
	UserHandler *users.UserHandle
	Wg          *sync.WaitGroup
}

func (a *App) Serve(l *Logger, cfg *Config, dbPool *pgxpool.Pool, ctx context.Context) error {
	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      a.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	shutdownError := make(chan error)
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		s := <-stop
		l.Info("Shuting down server", map[string]string{
			"signal": s.String(),
		})

		err := server.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}
		l.Info("Completing background tasks", map[string]string{
			"add": server.Addr,
		})
		a.Wg.Wait()

		shutdownError <- nil
	}()

	l.Info("Starting server", map[string]string{
		"add": server.Addr,
	})

	log.Printf("librorum listening on %s", cfg.Addr)
	log.Printf("database url configured: %t", os.Getenv("LIBRORUM_DATABASE_URL") != "")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("http server: %v", err)
	}
	err := <-shutdownError
	if err != nil {
		return err
	}

	l.Info("stopped server", map[string]string{
		"addr": server.Addr,
	})
	return nil
}

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", a.handleHealth)
	mux.HandleFunc("GET /api/example/library-items", a.BookHandler.DisplayBooks)
	mux.HandleFunc("POST /users/register", a.UserHandler.Register)
	return logRequests(mux)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func (a *App) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json response: %v", err)
	}
}
