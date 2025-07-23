package config

import (
	"os"
)

type Config struct {
	ProjectID     string
	Region        string
	GcsBucket     string
	DB            string // full postgres DSN
	FirebaseCreds string // path to service‑account JSON
}

func Load() *Config {
	return &Config{
		ProjectID:  os.Getenv("GCP_PROJECT"),
		Region:     os.Getenv("REGION"),
		GcsBucket:  os.Getenv("GCS_BUCKET"),
		DB:         os.Getenv("DB_DSN"),
		FirebaseCreds: os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
	}
}
