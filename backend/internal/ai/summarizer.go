package ai

import (
	"context"
	"log"

	"cloud.google.com/go/vertexai/genai"

	"github.com/hi-wesley/mini-youtube/internal/config"
	"github.com/hi-wesley/mini-youtube/internal/db"
	"github.com/hi-wesley/mini-youtube/internal/models"
)

func GenerateAndCacheSummary(videoID, gcsURI string) {
	ctx := context.Background()
	cfg := config.Load()

	client, err := genai.NewClient(ctx, cfg.ProjectID, cfg.Region)
	if err != nil {
		log.Printf("Failed to create genai client: %v", err)
		return
	}
	defer client.Close()

	modelName := "gemini-2.5-flash-lite" // latest model that understands video
	model := client.GenerativeModel(modelName)

	resp, err := model.GenerateContent(ctx, genai.FileData{MIMEType: "video/mp4", FileURI: gcsURI}, genai.Text("Summarize this video in 3 concise sentences."))
	if err != nil {
		log.Println(err)
		return
	}

	summary := resp.Candidates[0].Content.Parts[0].(genai.Text)
	db.Conn.Model(&models.Video{}).
		Where("id = ?", videoID).
		Updates(map[string]interface{}{
			"summary":       string(summary),
			"summary_model": modelName,
		})
}
