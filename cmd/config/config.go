package config

import (
	"encoding/hex"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr              string
	DataDir           string
	OLContact         string
	GoogleBooksAPIKey string
	DB                struct {
		host           string
		port           string
		username       string
		password       string
		name           string
		sslMode        string
		maxConnections int
		maxIdleTime    int
	}
	TokenExpiration struct {
		durationString string
		duration       time.Duration
	}
	Secret struct {
		HMC               string
		SecretKey         []byte
		SessionExpiration time.Duration
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
	flag.StringVar(&cfg.DB.host, "db-host", os.Getenv("DB_HOST"), "PostgreSQL host")
	flag.StringVar(&cfg.DB.port, "db-port", os.Getenv("DB_PORT"), "PostgreSQL port")
	flag.StringVar(&cfg.DB.username, "db-username", os.Getenv("DB_USER"), "PostgreSQL username")
	cfg.DB.password = os.Getenv("DB_PASSWORD")
	flag.StringVar(&cfg.DB.name, "db-name", os.Getenv("DB_NAME"), "PostgreSQL database name")
	flag.StringVar(&cfg.DB.sslMode, "db-ssl-mode", os.Getenv("DB_SSL_MODE"), "PostgreSQL SSL mode")
	flag.IntVar(&cfg.DB.maxIdleTime, "db-max-idle-time", maxIdleTime, "PostgreSQL max idle time")
	flag.StringVar(&cfg.DataDir, "data-dir", os.Getenv("LIBRORUM_DATA_DIR"), "Data directory")
	flag.StringVar(&cfg.Addr, "addr", os.Getenv("LIBRORUM_ADDR"), "Address")
	flag.StringVar(&cfg.OLContact, "ol-contact", os.Getenv("LIBRORUM_OPENLIBRARY_CONTACT"), "Open Library API contact")
	flag.StringVar(&cfg.GoogleBooksAPIKey, "google-books-api-key", os.Getenv("GOOGLE_BOOKS_API_KEY"), "Google Books API key")

	flag.StringVar(&cfg.Secret.HMC, "secret-key", os.Getenv("HMC_SECRET_KEY"), "HMC Secret Key")
	secretKey, err := hex.DecodeString(cfg.Secret.HMC)
	if err != nil {
		return nil, err
	}
	cfg.Secret.SecretKey = secretKey
	sessionDuration, err := time.ParseDuration(os.Getenv("SESSION_EXPIRATION"))
	if err != nil {
		return nil, err
	}
	cfg.Secret.SessionExpiration = sessionDuration

	tokexpirationStr := os.Getenv("TOKEN_EXPIRATION")
	duration, err := time.ParseDuration(tokexpirationStr)
	if err != nil {
		return nil, err
	}
	cfg.TokenExpiration.durationString = tokexpirationStr
	cfg.TokenExpiration.duration = duration
	flag.Parse()

	return &cfg, nil
}

func (cfg *Config) databaseDSN() (string, error) {
	if cfg.DB.host == "" || cfg.DB.port == "" || cfg.DB.username == "" || cfg.DB.password == "" || cfg.DB.name == "" {
		return "", fmt.Errorf("missing required database configuration")
	}

	databaseURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.DB.username, cfg.DB.password),
		Host:   net.JoinHostPort(cfg.DB.host, cfg.DB.port),
		Path:   "/" + cfg.DB.name,
	}
	if cfg.DB.sslMode != "" {
		query := databaseURL.Query()
		query.Set("sslmode", cfg.DB.sslMode)
		databaseURL.RawQuery = query.Encode()
	}

	return databaseURL.String(), nil
}

func (cfg *Config) databaseConfigured() bool {
	return cfg.DB.host != "" && cfg.DB.port != "" && cfg.DB.username != "" && cfg.DB.password != "" && cfg.DB.name != ""
}
