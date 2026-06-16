package config

import (
	"Librorum/internal/books"
	"Librorum/internal/users"
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
)

type App struct {
	BookHandler *books.Handler
	UserHandler *users.UserHandle
	Wg          *sync.WaitGroup
}

func (a *App) Serve(l *Logger, cfg *Config) error {
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

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

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
	//GET
	mux.HandleFunc("GET /healthz", a.handleHealth)
	mux.HandleFunc("GET /users/currentUser", a.UserHandler.CurrentUser)
	mux.HandleFunc("GET /books/library-items", a.BookHandler.DisplayBooks)
	mux.Handle("GET /covers/", http.StripPrefix("/covers/", http.FileServer(http.Dir(a.BookHandler.Paths.CoverCacheDir))))
	//POST
	mux.HandleFunc("POST /users/register", a.UserHandler.Register)
	mux.HandleFunc("POST /users/login", a.UserHandler.LoginUser)
	mux.HandleFunc("POST /users/logout", a.UserHandler.Logout)
	mux.HandleFunc("POST /books/insert", a.BookHandler.InsertEpubBooks)
	return logRequests(mux)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				start := time.Now()
				w.Header().Set("Connection", "close")
				log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
			}
		}()
		next.ServeHTTP(w, r)
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
