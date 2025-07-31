package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/hi-wesley/mini-youtube/internal/config"
	"github.com/hi-wesley/mini-youtube/internal/db"
	"github.com/hi-wesley/mini-youtube/internal/models"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type BackupData struct {
	Timestamp interface{}      `json:"timestamp"` // Can be string or time.Time
	Users     []models.User    `json:"users"`
	Videos    []models.Video   `json:"videos"`
	Comments  []models.Comment `json:"comments"`
	Likes     []models.Like    `json:"likes"`
	Firebase  []FirebaseUser   `json:"firebase_users"`
}

type FirebaseUser struct {
	UID         string `json:"uid"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run restore_all_data.go <backup_directory>")
		fmt.Println("Example: go run restore_all_data.go backups/backup_20240730_143022")
		os.Exit(1)
	}

	backupDir := os.Args[1]
	fmt.Printf("=== Restore Mini YouTube Data from %s ===\n", backupDir)
	fmt.Println("WARNING: This will DELETE ALL CURRENT DATA and replace it with the backup!")
	fmt.Print("Type 'RESTORE ALL' to confirm: ")
	
	var confirmation string
	fmt.Scanln(&confirmation)
	
	if confirmation != "RESTORE" {
		fmt.Println("Cancelled. No data was restored.")
		return
	}

	// Load .env file
	if err := godotenv.Load("../.env"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("Warning: Could not load .env file, using environment variables")
		}
	}

	// Load backup metadata
	metadataFile := filepath.Join(backupDir, "backup.json")
	data, err := os.ReadFile(metadataFile)
	if err != nil {
		log.Fatalf("Failed to read backup metadata: %v", err)
	}

	var backup BackupData
	if err := json.Unmarshal(data, &backup); err != nil {
		log.Fatalf("Failed to parse backup metadata: %v", err)
	}

	// Load config and connect to database
	cfg := config.Load()
	if err := db.Connect(cfg.DB); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	ctx := context.Background()

	// 1. Clear existing data
	fmt.Println("\n1. Clearing existing data...")
	if err := clearAllData(ctx, cfg); err != nil {
		log.Fatalf("Failed to clear existing data: %v", err)
	}

	// 2. Restore Firebase Authentication FIRST (to get UID mappings)
	fmt.Println("\n2. Restoring Firebase Authentication...")
	if err := restoreFirebaseAuth(ctx, backup.Firebase); err != nil {
		log.Printf("Error restoring Firebase Auth: %v", err)
	} else {
		fmt.Println("   ✓ Firebase Auth restored with UID mappings")
	}

	// 3. Restore Google Cloud Storage
	fmt.Println("\n3. Restoring Google Cloud Storage...")
	if err := restoreGoogleCloudStorage(ctx, cfg, backupDir); err != nil {
		log.Printf("Error restoring GCS: %v", err)
	} else {
		fmt.Println("   ✓ Google Cloud Storage restored")
	}

	// 4. Restore Supabase Database (using UID mappings from step 2)
	fmt.Println("\n4. Restoring Supabase database...")
	if err := restoreDatabase(backup); err != nil {
		log.Printf("Error restoring database: %v", err)
	} else {
		fmt.Println("   ✓ Database restored with remapped UIDs")
	}

	fmt.Printf("\n✅ Restore completed successfully!\n")
}

func clearAllData(ctx context.Context, cfg *config.Config) error {
	// Clear database
	tables := []string{"comments", "likes", "videos", "users"}
	for _, table := range tables {
		result := db.Conn.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if result.Error != nil {
			return fmt.Errorf("failed to clear %s table: %w", table, result.Error)
		}
		fmt.Printf("   - Cleared %s table\n", table)
	}

	// Clear Firebase Auth
	opt := option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Auth client: %w", err)
	}

	iter := client.Users(ctx, "")
	deletedCount := 0
	for {
		user, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("error iterating users: %w", err)
		}

		if err := client.DeleteUser(ctx, user.UID); err != nil {
			log.Printf("Error deleting user %s: %v", user.Email, err)
		} else {
			deletedCount++
		}
	}
	fmt.Printf("   - Deleted %d Firebase users\n", deletedCount)

	// Clear GCS
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}
	defer storageClient.Close()

	bucket := storageClient.Bucket(cfg.GcsBucket)
	iter2 := bucket.Objects(ctx, nil)
	deletedFiles := 0
	
	for {
		attrs, err := iter2.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("error iterating objects: %w", err)
		}

		if err := bucket.Object(attrs.Name).Delete(ctx); err != nil {
			log.Printf("Error deleting object %s: %v", attrs.Name, err)
		} else {
			deletedFiles++
		}
	}
	fmt.Printf("   - Deleted %d files from GCS\n", deletedFiles)

	return nil
}

func restoreDatabase(backup BackupData) error {
	// Restore users with mapped UIDs
	for _, user := range backup.Users {
		// Update user ID to new Firebase UID
		if newUID, ok := uidMapping[user.ID]; ok {
			user.ID = newUID
		} else {
			log.Printf("Warning: No UID mapping found for user %s (ID: %s)", user.Email, user.ID)
			continue
		}
		
		if err := db.Conn.Create(&user).Error; err != nil {
			log.Printf("Error creating user %s: %v", user.Email, err)
		}
	}
	fmt.Printf("   - Restored %d users with new UIDs\n", len(backup.Users))

	// Restore videos with mapped user IDs
	for _, video := range backup.Videos {
		// Update video owner ID to new Firebase UID
		if newUID, ok := uidMapping[video.UserID]; ok {
			video.UserID = newUID
		} else {
			log.Printf("Warning: No UID mapping found for video owner %s", video.UserID)
			continue
		}
		
		if err := db.Conn.Create(&video).Error; err != nil {
			log.Printf("Error creating video %s: %v", video.Title, err)
		}
	}
	fmt.Printf("   - Restored %d videos with mapped user IDs\n", len(backup.Videos))

	// Restore comments with mapped user IDs
	restoredComments := 0
	for _, comment := range backup.Comments {
		// Update comment author ID to new Firebase UID
		if newUID, ok := uidMapping[comment.UserID]; ok {
			comment.UserID = newUID
		} else {
			log.Printf("Warning: Skipping comment from unknown user %s", comment.UserID)
			continue
		}
		
		if err := db.Conn.Create(&comment).Error; err != nil {
			log.Printf("Error creating comment: %v", err)
		} else {
			restoredComments++
		}
	}
	fmt.Printf("   - Restored %d comments with mapped user IDs\n", restoredComments)

	// Restore likes with mapped user IDs
	restoredLikes := 0
	for _, like := range backup.Likes {
		// Update like user ID to new Firebase UID
		if newUID, ok := uidMapping[like.UserID]; ok {
			like.UserID = newUID
		} else {
			log.Printf("Warning: Skipping like from unknown user %s", like.UserID)
			continue
		}
		
		if err := db.Conn.Create(&like).Error; err != nil {
			log.Printf("Error creating like: %v", err)
		} else {
			restoredLikes++
		}
	}
	fmt.Printf("   - Restored %d likes with mapped user IDs\n", restoredLikes)

	return nil
}

// Global map to track old UID -> new UID mappings
var uidMapping = make(map[string]string)

func restoreFirebaseAuth(ctx context.Context, users []FirebaseUser) error {
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

	// Create new users and build UID mapping
	createdCount := 0
	for _, user := range users {
		params := (&auth.UserToCreate{}).
			Email(user.Email).
			EmailVerified(true)
		
		if user.DisplayName != "" {
			params = params.DisplayName(user.DisplayName)
		}

		newUser, err := client.CreateUser(ctx, params)
		if err != nil {
			log.Printf("Error creating user %s: %v", user.Email, err)
		} else {
			// Map old UID to new UID
			uidMapping[user.UID] = newUser.UID
			createdCount++
			fmt.Printf("   - Mapped user %s: %s -> %s\n", user.Email, user.UID, newUser.UID)
		}
	}

	fmt.Printf("   - Created %d Firebase users with UID mapping\n", createdCount)
	fmt.Println("   ⚠️  Note: Restored users must use 'Forgot Password' to set a new password")
	fmt.Println("   ⚠️  Admin can trigger password reset emails from Firebase Console")
	return nil
}

func restoreGoogleCloudStorage(ctx context.Context, cfg *config.Config, backupDir string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}
	defer client.Close()

	bucket := client.Bucket(cfg.GcsBucket)
	videosDir := filepath.Join(backupDir, "videos")

	// Walk through all files in the backup
	uploadedCount := 0
	err = filepath.Walk(videosDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Calculate GCS object name
		relPath, err := filepath.Rel(videosDir, path)
		if err != nil {
			return err
		}
		gcsPath := strings.ReplaceAll(relPath, string(os.PathSeparator), "/")

		// Open local file
		file, err := os.Open(path)
		if err != nil {
			log.Printf("Error opening file %s: %v", path, err)
			return nil
		}
		defer file.Close()

		// Upload to GCS
		wc := bucket.Object(gcsPath).NewWriter(ctx)
		if _, err := io.Copy(wc, file); err != nil {
			log.Printf("Error uploading %s: %v", gcsPath, err)
			return nil
		}
		if err := wc.Close(); err != nil {
			log.Printf("Error closing writer for %s: %v", gcsPath, err)
			return nil
		}

		uploadedCount++
		fmt.Printf("   - Uploaded: %s\n", gcsPath)
		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking backup directory: %w", err)
	}

	fmt.Printf("   - Total files uploaded: %d\n", uploadedCount)
	return nil
}