package db

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/<org>/mini-youtube/internal/models"
)

var Conn *gorm.DB

func Connect(dsn string) error {
	var err error
	Conn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	return err
}

func AutoMigrate() error {
	return Conn.AutoMigrate(&models.User{}, &models.Video{},
		&models.Comment{}, &models.Like{})
}

// common helper
func Paginator(page, pageSize int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 { page = 1 }
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
