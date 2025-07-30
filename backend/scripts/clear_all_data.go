package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"github.com/hi-wesley/mini-youtube/internal/config"
	"github.com/hi-wesley/mini-youtube/internal/db"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func main() {
	fmt.Println("=== Mini YouTube Data Cleanup Script ===")
	fmt.Println("This will DELETE ALL DATA from:")
	fmt.Println("1. Supabase (all videos, users, comments)")
	fmt.Println("2. Firebase Authentication (all user accounts)")
	fmt.Println("3. Google Cloud Storage (all video files)")
	fmt.Println()
	fmt.Print("Are you ABSOLUTELY SURE? Type 'DELETE ALL' to confirm: ")
	
	reader := bufio.NewReader(os.Stdin)
	confirmation, _ := reader.ReadString('\n')
	confirmation = strings.TrimSpace(confirmation)
	
	if confirmation != "DELETE ALL" {
		fmt.Printf("You typed: '%s'\n", confirmation)
		fmt.Println("Cancelled. No data was deleted.")
		return
	}

	// Load .env file from parent directory (since we're in scripts/)
	if err := godotenv.Load("../.env"); err != nil {
		// Try current directory as fallback
		if err := godotenv.Load(".env"); err != nil {
			log.Println("Warning: Could not load .env file, using environment variables")
		}
	}
	
	// Load config
	cfg := config.Load()

	ctx := context.Background()

	// 1. Clear Supabase Database
	fmt.Println("\n1. Clearing Supabase database...")
	if err := clearDatabase(cfg); err != nil {
		log.Printf("Error clearing database: %v", err)
	} else {
		fmt.Println("   ✓ Database cleared")
	}

	// 2. Clear Firebase Authentication
	fmt.Println("\n2. Clearing Firebase Authentication...")
	if err := clearFirebaseAuth(ctx); err != nil {
		log.Printf("Error clearing Firebase Auth: %v", err)
	} else {
		fmt.Println("   ✓ Firebase Auth cleared")
	}

	// 3. Clear Google Cloud Storage
	fmt.Println("\n3. Clearing Google Cloud Storage...")
	if err := clearGoogleCloudStorage(ctx, cfg); err != nil {
		log.Printf("Error clearing GCS: %v", err)
	} else {
		fmt.Println("   ✓ Google Cloud Storage cleared")
	}

	fmt.Println("\n✅ All data has been cleared successfully!")
}

func clearDatabase(cfg *config.Config) error {
	if err := db.Connect(cfg.DB); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	database := db.Conn

	// Delete in correct order due to foreign key constraints
	tables := []string{"comments", "videos", "users"}
	
	for _, table := range tables {
		result := database.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if result.Error != nil {
			return fmt.Errorf("failed to clear %s table: %w", table, result.Error)
		}
		fmt.Printf("   - Deleted %d records from %s\n", result.RowsAffected, table)
	}

	return nil
}

func clearFirebaseAuth(ctx context.Context) error {
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

	// List and delete all users
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
			log.Printf("   - Error deleting user %s: %v", user.Email, err)
		} else {
			deletedCount++
		}
	}

	fmt.Printf("   - Deleted %d users from Firebase Auth\n", deletedCount)
	return nil
}

func clearGoogleCloudStorage(ctx context.Context, cfg *config.Config) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}
	defer client.Close()

	bucket := client.Bucket(cfg.GcsBucket)
	
	// List and delete all objects
	iter := bucket.Objects(ctx, nil)
	deletedCount := 0
	
	for {
		attrs, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("error iterating objects: %w", err)
		}

		if err := bucket.Object(attrs.Name).Delete(ctx); err != nil {
			log.Printf("   - Error deleting object %s: %v", attrs.Name, err)
		} else {
			deletedCount++
		}
	}

	fmt.Printf("   - Deleted %d objects from GCS bucket %s\n", deletedCount, cfg.GcsBucket)
	return nil
}