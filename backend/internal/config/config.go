package config

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	ProjectID     string
	Region        string
	GcsBucket     string
	DB            string // full postgres DSN
	FirebaseCreds string // path to service‑account JSON
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
			ProjectID:     os.Getenv("GCP_PROJECT"),
			Region:        os.Getenv("REGION"),
			GcsBucket:     os.Getenv("GCS_BUCKET"),
			DB:            os.Getenv("DB_DSN"),
			FirebaseCreds: os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		}
	})
	return cfg
}
