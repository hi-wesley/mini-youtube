package main

import (
	"context"
	"log"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/iterator"

	"github.com/hi-wesley/mini-youtube/internal/config"
	"github.com/hi-wesley/mini-youtube/internal/db"
)

func main() {
	log.Println("--- WARNING: This is a destructive operation. ---")
	log.Println("This script will delete all data from GCS, Supabase, and Firebase Auth.")

	// ----- Load Configuration -----
	cfg := config.Load()
	ctx := context.Background()

	// ----- 1. Clear Google Cloud Storage -----
	log.Println("\n--- Clearing Google Cloud Storage ---")
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create GCS client: %v", err)
	}
	defer storageClient.Close()

	bucket := storageClient.Bucket(cfg.GcsBucket)
	it := bucket.Objects(ctx, nil)
	objCount := 0
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate GCS objects: %v", err)
		}
		if err := bucket.Object(attrs.Name).Delete(ctx); err != nil {
			log.Printf("Failed to delete object %s: %v", attrs.Name, err)
		} else {
			log.Printf("Deleted GCS object: %s", attrs.Name)
			objCount++
		}
	}
	log.Printf("Successfully deleted %d objects from bucket '%s'.", objCount, cfg.GcsBucket)

	// ----- 2. Clear Supabase Database -----
	log.Println("\n--- Clearing Supabase Database ---")
	if err := db.Connect(cfg.DB); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// The order is important due to foreign key constraints.
	// We delete from tables that are depended upon last.
	tables := []string{"likes", "comments", "videos", "users"}
	for _, table := range tables {
		log.Printf("Truncating table: %s", table)
		if err := db.Conn.Exec("TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE;").Error; err != nil {
			log.Fatalf("Failed to truncate table %s: %v", table, err)
		}
		log.Printf("Successfully truncated '%s' table.", table)
	}
	log.Println("Successfully cleared all tables from the database.")

	// ----- 3. Clear Firebase Authentication -----
	log.Println("\n--- Clearing Firebase Authentication Users ---")
	
	// Initialize Firebase Admin SDK
	// Note: This assumes GOOGLE_APPLICATION_CREDENTIALS is set in your environment.
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase app: %v", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("Failed to create Firebase Auth client: %v", err)
	}

	userIt := client.Users(ctx, "")
	userCount := 0
	for {
		user, err := userIt.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate Firebase users: %v", err)
		}
		if err := client.DeleteUser(ctx, user.UID); err != nil {
			log.Printf("Failed to delete user %s (%s): %v", user.UID, user.Email, err)
		} else {
			log.Printf("Deleted Firebase user: %s (%s)", user.UID, user.Email)
			userCount++
		}
	}
	log.Printf("Successfully deleted %d users from Firebase Authentication.", userCount)

	log.Println("\n--- Cleanup Complete ---")
}
