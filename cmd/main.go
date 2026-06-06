package main

import (
	"context"
	"log"
	"os"
	"time"

	"Librorum/cmd/config"
	logger "Librorum/cmd/config"
	"Librorum/internal/books"
	db "Librorum/internal/platform/storage/sqlc"
	"Librorum/internal/storage"
	users "Librorum/internal/users/register"
)

func main() {
	appCtx := context.Background()
	setupCtx, cancel := context.WithTimeout(appCtx, 10*time.Second)
	defer cancel()

	logger := logger.New(os.Stdout, logger.LevelInfo)

	cfg, _ := config.LoadConfig(logger)

	dbPool, err := config.OpenDB(logger, cfg, setupCtx)
	if err != nil || dbPool.Ping(setupCtx) != nil {
		logger.Fatal("Failed to connect to DB: %v"+err.Error(), nil)
	}

	defer dbPool.Close()

	paths := storage.NewPaths(cfg.DataDir)
	if err := storage.EnsureDirs(paths); err != nil {
		logger.Fatal("prepare storage directories: %v"+err.Error(), nil)
	}

	h := &books.Handler{
		Db:      dbPool,
		Queries: db.New(dbPool),
	}
	u := &users.UserHandle{
		DB:      dbPool,
		Queries: db.New(dbPool),
	}
	app := &config.App{
		BookHandler: h,
		UserHandler: u,
	}

	app.Serve(logger, cfg, dbPool, setupCtx)

	log.Print("librorum shutting down")
}
