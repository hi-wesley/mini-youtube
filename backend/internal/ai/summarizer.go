package ai

import (
	"context"
	"log"

	genai "cloud.google.com/go/vertexai/genai/v1"
	"cloud.google.com/go/vertexai/vertexai"
	"google.golang.org/api/option"

	"github.com/<org>/mini-youtube/internal/db"
	"github.com/<org>/mini-youtube/internal/models"
)

func GenerateAndCacheSummary(videoID, gcsURI string) {
	ctx := context.Background()
	cfg := config.Load()

	client, err := genai.NewClient(ctx,
		option.WithEndpoint(fmt.Sprintf("%s-aiplatform.googleapis.com:443", cfg.Region)))
	if err != nil { log.Println(err); return }
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-pro") // latest model that understands video :contentReference[oaicite:1]{index=1}

	resp, err := model.GenerateContent(ctx, &genai.GenerateContentRequest{
		Contents: []*genai.Content{{
			Parts: []genai.Part{
				genai.FileDataPart("video/mp4", gcsURI),
			},
		}, {
			Parts: []genai.Part{
				genai.TextPart("Summarize this video in 3 concise sentences."),
			},
		}},
	})
	if err != nil { log.Println(err); return }

	summary := resp.GetText()
	db.Conn.Model(&models.Video{}).
		Where("id = ?", videoID).
		Update("summary", summary)
}
