// This file manages the connection to the database.
// It provides a function to connect to the PostgreSQL database using the
// connection string provided by the config file.
package db

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/hi-wesley/mini-youtube/internal/models"
)

var Conn *gorm.DB

func Connect(dsn string) error {
	var err error
	log.Printf("Connecting to database with DSN: %s", dsn)
	Conn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: false,
	})
	return err
}

func AutoMigrate() error {
	return Conn.AutoMigrate(&models.User{}, &models.Video{},
		&models.Comment{}, &models.Like{})
}

// common helper
func Paginator(page, pageSize int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}
