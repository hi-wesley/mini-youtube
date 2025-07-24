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
		log.Println(err)
		return
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-pro") // latest model that understands video

	resp, err := model.GenerateContent(ctx, genai.FileData{MIMEType: "video/mp4", FileURI: gcsURI}, genai.Text("Summarize this video in 3 concise sentences."))
	if err != nil {
		log.Println(err)
		return
	}

	summary := resp.Candidates[0].Content.Parts[0].(genai.Text)
	db.Conn.Model(&models.Video{}).
		Where("id = ?", videoID).
		Update("summary", summary)
}
