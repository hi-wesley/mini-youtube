package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"github.com/hi-wesley/mini-youtube/internal/config"
	"github.com/hi-wesley/mini-youtube/internal/db"
	"github.com/hi-wesley/mini-youtube/internal/models"
	"gorm.io/gorm"
)

func main() {
	// 1. Get Video ID from command-line arguments
	if len(os.Args) < 2 {
		log.Fatal("Please provide a video ID to delete.")
	}
	videoID := os.Args[1]

	// 2. Load configuration and connect to the database
	cfg := config.Load()
	if err := db.Connect(cfg.DB); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 3. Find the video in the database
	var video models.Video
	if err := db.Conn.First(&video, "id = ?", videoID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Fatalf("No video found with ID: %s", videoID)
		}
		log.Fatalf("Failed to find video: %v", err)
	}
	log.Printf("Found video: %s (Title: %s)", video.ID, video.Title)

	// 4. Delete the video file from Google Cloud Storage
	log.Printf("Attempting to delete GCS object: %s", video.ObjectName)
	ctx := context.Background()
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create GCS client: %v", err)
	}
	defer storageClient.Close()

	bucket := storageClient.Bucket(cfg.GcsBucket)
	object := bucket.Object(video.ObjectName)
	if err := object.Delete(ctx); err != nil {
		log.Printf("Warning: Failed to delete GCS object '%s'. You may need to delete it manually. Error: %v", video.ObjectName, err)
	} else {
		log.Printf("Successfully deleted GCS object: %s", video.ObjectName)
	}

	// 5. Delete records from the database in a transaction
	log.Println("Attempting to delete database records...")
	err = db.Conn.Transaction(func(tx *gorm.DB) error {
		// Delete likes
		if err := tx.Where("video_id = ?", videoID).Delete(&models.Like{}).Error; err != nil {
			return err
		}
		// Delete comments
		if err := tx.Where("video_id = ?", videoID).Delete(&models.Comment{}).Error; err != nil {
			return err
		}
		// Delete video
		if err := tx.Where("id = ?", videoID).Delete(&models.Video{}).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Failed to delete database records: %v", err)
	}

	log.Println("Successfully deleted all database records for the video.")
	fmt.Printf("✅ Video with ID %s has been completely deleted.\n", videoID)
}