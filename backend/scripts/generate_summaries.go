package main

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/vertexai/genai"
	"github.com/hi-wesley/mini-youtube/internal/config"
	"github.com/hi-wesley/mini-youtube/internal/db"
	"github.com/hi-wesley/mini-youtube/internal/models"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("=== Generate Summaries for Existing Videos ===")

	// Load .env file
	if err := godotenv.Load("../.env"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("Warning: Could not load .env file, using environment variables")
		}
	}

	// Load config and connect to database
	cfg := config.Load()
	if err := db.Connect(cfg.DB); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	ctx := context.Background()

	// Create Vertex AI client
	client, err := genai.NewClient(ctx, cfg.ProjectID, cfg.Region)
	if err != nil {
		log.Fatalf("Failed to create genai client: %v", err)
		return
	}
	defer client.Close()

	// Get all videos (regenerate summaries for all)
	var videos []models.Video
	if err := db.Conn.Find(&videos).Error; err != nil {
		log.Fatalf("Failed to fetch videos: %v", err)
	}

	fmt.Printf("Found %d videos total. Regenerating summaries for all.\n", len(videos))

	modelName := "gemini-2.5-pro"
	model := client.GenerativeModel(modelName)

	// Process each video
	for i, video := range videos {
		fmt.Printf("\n[%d/%d] Processing video: %s\n", i+1, len(videos), video.Title)
		
		gcsURI := fmt.Sprintf("gs://%s/%s", cfg.GcsBucket, video.ObjectName)
		fmt.Printf("  GCS URI: %s\n", gcsURI)

		// Generate summary
		resp, err := model.GenerateContent(ctx, 
			genai.FileData{
				MIMEType: "video/mp4", 
				FileURI: gcsURI,
			}, 
			genai.Text("Summarize this video in 3 concise sentences."),
		)
		
		if err != nil {
			log.Printf("  ERROR generating summary: %v", err)
			
			// Try to get more details about the error
			fmt.Printf("  Full error details: %+v\n", err)
			
			// Common issues:
			// - File not found in GCS
			// - Permissions issue
			// - Region mismatch
			// - API not enabled
			continue
		}

		// Extract summary from response
		if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
			summary := resp.Candidates[0].Content.Parts[0].(genai.Text)
			
			// Update database
			if err := db.Conn.Model(&models.Video{}).
				Where("id = ?", video.ID).
				Updates(map[string]interface{}{
					"summary":       string(summary),
					"summary_model": modelName,
				}).Error; err != nil {
				log.Printf("  ERROR updating database: %v", err)
			} else {
				fmt.Printf("  SUCCESS: Summary generated and saved\n")
				fmt.Printf("  Summary: %s\n", summary)
			}
		} else {
			fmt.Printf("  ERROR: No summary in response\n")
		}
	}

	fmt.Println("\nDone!")
}

// Debug function to test a single video
func testSingleVideo(ctx context.Context, client *genai.Client, videoID string) {
	var video models.Video
	if err := db.Conn.Where("id = ?", videoID).First(&video).Error; err != nil {
		log.Printf("Video not found: %v", err)
		return
	}

	cfg := config.Load()
	gcsURI := fmt.Sprintf("gs://%s/%s", cfg.GcsBucket, video.ObjectName)
	
	fmt.Printf("Testing video: %s\n", video.Title)
	fmt.Printf("GCS URI: %s\n", gcsURI)
	
	model := client.GenerativeModel("gemini-2.5-flash")
	
	resp, err := model.GenerateContent(ctx, 
		genai.FileData{
			MIMEType: "video/mp4", 
			FileURI: gcsURI,
		}, 
		genai.Text("Summarize this video in 3 concise sentences."),
	)
	
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		return
	}
	
	fmt.Printf("Response: %+v\n", resp)
}