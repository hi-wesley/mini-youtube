package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"github.com/hi-wesley/mini-youtube/internal/config"
	"github.com/hi-wesley/mini-youtube/internal/db"
	"github.com/hi-wesley/mini-youtube/internal/models"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type BackupData struct {
	Timestamp time.Time              `json:"timestamp"`
	Users     []models.User          `json:"users"`
	Videos    []models.Video         `json:"videos"`
	Comments  []models.Comment       `json:"comments"`
	Likes     []models.Like          `json:"likes"`
	Firebase  []FirebaseUser         `json:"firebase_users"`
}

type FirebaseUser struct {
	UID         string `json:"uid"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

func main() {
	fmt.Println("=== Backup Mini YouTube Data ===")

	// Load .env file
	if err := godotenv.Load("../.env"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("Warning: Could not load .env file, using environment variables")
		}
	}

	// Create backup directory with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupDir := fmt.Sprintf("backups/backup_%s", timestamp)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		log.Fatalf("Failed to create backup directory: %v", err)
	}

	// Load config and connect to database
	cfg := config.Load()
	if err := db.Connect(cfg.DB); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	ctx := context.Background()
	backup := BackupData{Timestamp: time.Now()}

	// 1. Backup Supabase Database
	fmt.Println("\n1. Backing up Supabase database...")
	if err := backupDatabase(&backup); err != nil {
		log.Printf("Error backing up database: %v", err)
	} else {
		fmt.Println("   ✓ Database backed up")
	}

	// 2. Backup Firebase Authentication
	fmt.Println("\n2. Backing up Firebase Authentication...")
	if err := backupFirebaseAuth(ctx, &backup); err != nil {
		log.Printf("Error backing up Firebase Auth: %v", err)
	} else {
		fmt.Println("   ✓ Firebase Auth backed up")
	}

	// Save backup metadata
	metadataFile := filepath.Join(backupDir, "backup.json")
	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal backup data: %v", err)
	}
	if err := os.WriteFile(metadataFile, data, 0644); err != nil {
		log.Fatalf("Failed to write metadata file: %v", err)
	}

	// 3. Backup Google Cloud Storage
	fmt.Println("\n3. Backing up Google Cloud Storage...")
	if err := backupGoogleCloudStorage(ctx, cfg, backupDir); err != nil {
		log.Printf("Error backing up GCS: %v", err)
	} else {
		fmt.Println("   ✓ Google Cloud Storage backed up")
	}

	fmt.Printf("\n✅ Backup completed successfully!\n")
	fmt.Printf("📁 Backup saved to: %s\n", backupDir)
}

func backupDatabase(backup *BackupData) error {
	// Backup users
	if err := db.Conn.Find(&backup.Users).Error; err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}
	fmt.Printf("   - Backed up %d users\n", len(backup.Users))

	// Backup videos
	if err := db.Conn.Find(&backup.Videos).Error; err != nil {
		return fmt.Errorf("failed to fetch videos: %w", err)
	}
	fmt.Printf("   - Backed up %d videos\n", len(backup.Videos))

	// Backup comments
	if err := db.Conn.Find(&backup.Comments).Error; err != nil {
		return fmt.Errorf("failed to fetch comments: %w", err)
	}
	fmt.Printf("   - Backed up %d comments\n", len(backup.Comments))

	// Backup likes
	if err := db.Conn.Find(&backup.Likes).Error; err != nil {
		return fmt.Errorf("failed to fetch likes: %w", err)
	}
	fmt.Printf("   - Backed up %d likes\n", len(backup.Likes))

	return nil
}

func backupFirebaseAuth(ctx context.Context, backup *BackupData) error {
	// Initialize Firebase Admin SDK
	opt := option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Auth client: %w", err)
	}

	// List all users
	iter := client.Users(ctx, "")
	userCount := 0
	
	for {
		user, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("error iterating users: %w", err)
		}

		backup.Firebase = append(backup.Firebase, FirebaseUser{
			UID:         user.UID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
		})
		userCount++
	}

	fmt.Printf("   - Backed up %d Firebase users\n", userCount)
	return nil
}

func backupGoogleCloudStorage(ctx context.Context, cfg *config.Config, backupDir string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}
	defer client.Close()

	bucket := client.Bucket(cfg.GcsBucket)
	videosDir := filepath.Join(backupDir, "videos")
	if err := os.MkdirAll(videosDir, 0755); err != nil {
		return fmt.Errorf("failed to create videos directory: %w", err)
	}

	// List and download all objects
	iter := bucket.Objects(ctx, nil)
	downloadedCount := 0
	
	for {
		attrs, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("error iterating objects: %w", err)
		}

		// Create local file path preserving GCS structure
		localPath := filepath.Join(videosDir, attrs.Name)
		localDir := filepath.Dir(localPath)
		if err := os.MkdirAll(localDir, 0755); err != nil {
			log.Printf("Error creating directory %s: %v", localDir, err)
			continue
		}

		// Download the object
		rc, err := bucket.Object(attrs.Name).NewReader(ctx)
		if err != nil {
			log.Printf("Error reading object %s: %v", attrs.Name, err)
			continue
		}

		file, err := os.Create(localPath)
		if err != nil {
			rc.Close()
			log.Printf("Error creating file %s: %v", localPath, err)
			continue
		}

		if _, err := io.Copy(file, rc); err != nil {
			rc.Close()
			file.Close()
			log.Printf("Error downloading %s: %v", attrs.Name, err)
			continue
		}

		rc.Close()
		file.Close()
		downloadedCount++
		fmt.Printf("   - Downloaded: %s\n", attrs.Name)
	}

	fmt.Printf("   - Total files downloaded: %d\n", downloadedCount)
	return nil
}