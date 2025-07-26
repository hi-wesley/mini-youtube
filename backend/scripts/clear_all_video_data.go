package main

import (
	"log"

	"github.com/hi-wesley/mini-youtube/internal/config"
	"github.com/hi-wesley/mini-youtube/internal/db"
)

func main() {
	cfg := config.Load()
	if err := db.Connect(cfg.DB); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Clearing all video-related data from the database...")

	// Truncate tables in correct order to avoid foreign key constraints
	// CASCADE is used to also truncate any tables that have foreign key constraints
	// pointing to these tables, and RESTART IDENTITY resets the primary key sequences.
	
	// Note: The order of truncation is important if there are foreign key constraints
	// between these tables. Likes and Comments depend on Videos, so they should be
	// truncated first, or use CASCADE.

	if err := db.Conn.Exec("TRUNCATE TABLE likes RESTART IDENTITY CASCADE;").Error; err != nil {
		log.Fatalf("Failed to truncate likes table: %v", err)
	}
	log.Println("Truncated 'likes' table.")

	if err := db.Conn.Exec("TRUNCATE TABLE comments RESTART IDENTITY CASCADE;").Error; err != nil {
		log.Fatalf("Failed to truncate comments table: %v", err)
	}
	log.Println("Truncated 'comments' table.")

	if err := db.Conn.Exec("TRUNCATE TABLE videos RESTART IDENTITY CASCADE;").Error; err != nil {
		log.Fatalf("Failed to truncate videos table: %v", err)
	}
	log.Println("Truncated 'videos' table.")

	log.Println("Successfully cleared all video-related data from the database.")
}