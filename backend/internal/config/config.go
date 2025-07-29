// This file handles the application's configuration.
// It reads important settings, like database connection strings and security keys,
// from environment variables or a local `.env` file. This keeps sensitive
// information separate from the main codebase.
package config

import (
	"log"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	ProjectID      string
	Region         string
	GcsBucket      string
	DB             string // full postgres DSN
	FirebaseCreds  string // path to service‑account JSON
	AllowedOrigins string
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

		cfg = &Config{
			ProjectID:      os.Getenv("GCP_PROJECT"),
			Region:         os.Getenv("REGION"),
			GcsBucket:      "mini-youtube",
			DB:             strings.Trim(os.Getenv("DB_DSN"), `"`),
			FirebaseCreds:  os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
			AllowedOrigins: os.Getenv("ALLOWED_ORIGINS"),
		}

		
		if cfg.ProjectID == "" {
			log.Fatal("GCP_PROJECT environment variable is required")
		}
		if cfg.DB == "" {
			log.Fatal("DB_DSN environment variable is required")
		}
	})
	return cfg
}