package main

import (
	"context"
	"encoding/gob"
	"os"
	"sync"
	"time"

	"Librorum/cmd/config"
	logger "Librorum/cmd/config"
	"Librorum/internal/books"
	db "Librorum/internal/platform/storage/sqlc"
	"Librorum/internal/storage"
	users "Librorum/internal/users"
)

func main() {
	appCtx := context.Background()
	setupCtx, cancel := context.WithTimeout(appCtx, 10*time.Second)
	defer cancel()

	logger := logger.New(os.Stdout, logger.LevelInfo)

	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.Fatal("Failed to Load the configuration of the server: "+err.Error(), nil)
	}

	dbPool, err := config.OpenDB(logger, cfg, setupCtx)
	if err != nil {
		logger.Fatal("Failed to connect to DB: "+err.Error(), nil)
	}
	if dbPool.Ping(setupCtx) != nil {
		logger.Fatal("Failed to ping the DB: "+dbPool.Ping(setupCtx).Error(), nil)
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
		DB:            dbPool,
		Queries:       db.New(dbPool),
		Logger:        logger,
		SessionConfig: cfg.Secret,
	}
	app := &config.App{
		BookHandler: h,
		UserHandler: u,
		Wg:          &sync.WaitGroup{},
	}
	gob.Register(&u.User.Id)

	app.Serve(logger, cfg)

	logger.Info("librorum shutting down", nil)
}
