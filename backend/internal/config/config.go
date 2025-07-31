// This file handles the application's configuration.
// It reads important settings, like database connection strings and security keys,
// from environment variables or a local `.env` file. This keeps sensitive
// information separate from the main codebase.
package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	ProjectID          string
	Region             string
	GcsBucket          string
	DB                 string // full postgres DSN
	FirebaseCreds      string // path to service‑account JSON
	AllowedOrigins     string
	RateLimitEnabled   bool
	RateLimitRedisURL  string
	RateLimitRedisDB   int
}

var (
	cfg  *Config
	once sync.Once
)

func Load() *Config {
	once.Do(func() {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using environment variables")
		}

		redisDB := 0
		if dbStr := os.Getenv("RATE_LIMIT_REDIS_DB"); dbStr != "" {
			if db, err := strconv.Atoi(dbStr); err == nil {
				redisDB = db
			}
		}

		cfg = &Config{
			ProjectID:         os.Getenv("GCP_PROJECT"),
			Region:            os.Getenv("REGION"),
			GcsBucket:         os.Getenv("GCS_BUCKET"),
			DB:                strings.Trim(os.Getenv("DB_DSN"), `"`),
			FirebaseCreds:     os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
			AllowedOrigins:    os.Getenv("ALLOWED_ORIGINS"),
			RateLimitEnabled:  os.Getenv("RATE_LIMIT_ENABLED") == "true",
			RateLimitRedisURL: os.Getenv("RATE_LIMIT_REDIS_URL"),
			RateLimitRedisDB:  redisDB,
		}

		
		if cfg.ProjectID == "" {
			log.Fatal("GCP_PROJECT environment variable is required")
		}
		if cfg.DB == "" {
			log.Fatal("DB_DSN environment variable is required")
		}
		if cfg.GcsBucket == "" {
			log.Fatal("GCS_BUCKET environment variable is required")
		}
	})
	return cfg
}