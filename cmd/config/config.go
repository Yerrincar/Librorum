package config

import (
	"flag"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr      string
	DataDir   string
	OLContact string
	DB        struct {
		dsn            string
		maxConnections int
		maxIdleTime    int
	}
}

func LoadConfig(l *Logger) (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		l.Fatal("Error trying to Load .env: "+err.Error(), nil)
	}

	maxOpenConnsStr := os.Getenv("DB_MAX_OPEN_CONNS")
	maxOpenConns, err := strconv.Atoi(maxOpenConnsStr)
	if err != nil {
		l.Fatal("Error trying to Read DB_MAX_OPEN_CONNS from .env %v"+err.Error(), nil)
	}

	maxIdleTimeStr := os.Getenv("DB_MAX_IDLE_TIME")
	maxIdleTime, err := strconv.Atoi(maxIdleTimeStr)
	if err != nil {
		l.Fatal("Error trying to Read DB_MAX_IDLE_CONNS from .env %v"+err.Error(), nil)
	}
	var cfg Config

	flag.IntVar(&cfg.DB.maxConnections, "db-max-open-conns", maxOpenConns, "PostgreSQL max open connections")
	flag.StringVar(&cfg.DB.dsn, "db-dns", os.Getenv("LIBRORUM_DATABASE_URL"), "PostgreSQL DSN")
	flag.IntVar(&cfg.DB.maxIdleTime, "db-max-idle-time", maxIdleTime, "PostgreSQL max idle time")
	flag.StringVar(&cfg.DataDir, "data-dir", os.Getenv("LIBRORUM_DATA_DIR"), "Data directory")
	flag.StringVar(&cfg.Addr, "addr", os.Getenv("LIBRORUM_ADDR"), "Address")
	flag.StringVar(&cfg.OLContact, "ol-contact", os.Getenv("LIBRORUM_OPENLIBRARY_CONTACT"), "Open Library API contact")

	flag.Parse()

	return &cfg, nil
}
