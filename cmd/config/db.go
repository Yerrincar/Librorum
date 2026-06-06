package config

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func OpenDB(l *Logger, cfg *Config, ctx context.Context) (*pgxpool.Pool, error) {
	db, err := pgxpool.NewWithConfig(ctx, DBConfig(l, cfg))
	if err != nil {
		l.Fatal("Error trying to connect to the DB: %v"+err.Error(), nil)
	}
	return db, nil
}

func DBConfig(l *Logger, cfg *Config) *pgxpool.Config {
	dbConfig, err := pgxpool.ParseConfig(cfg.DB.dsn)
	if err != nil {
		l.Fatal("Error trying to parse DB config: %v"+err.Error(), nil)
	}

	dbConfig.MaxConns = int32(cfg.DB.maxConnections)
	dbConfig.MaxConnIdleTime = time.Duration(cfg.DB.maxIdleTime * int(time.Second))

	return dbConfig
}
