// This file contains the "handlers" for all video-related actions.
// A handler is a function that takes an incoming web request and performs
// the necessary actions. For example, it handles uploading videos, fetching
// a list of videos, or getting the details for a single video.
package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hi-wesley/mini-youtube/internal/ai"
	"github.com/hi-wesley/mini-youtube/internal/config"
	"github.com/hi-wesley/mini-youtube/internal/db"
	"github.com/hi-wesley/mini-youtube/internal/models"
	"github.com/modfy/fluent-ffmpeg"
	"gorm.io/gorm"
)

var (
	storageClient *storage.Client
	cfg           *config.Config
)

func init() {
	var err error
	storageClient, err = storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create storage client: %v", err)
	}
	cfg = config.Load()
}

// InitiateUpload generates a signed URL for direct GCS upload.
func InitiateUpload(c *gin.Context) {
	uid := c.GetString("uid")
	var req struct {
		FileName string `json:"fileName" binding:"required"`
		FileType string `json:"fileType" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fileName and fileType are required"})
		return
	}

	objectName := fmt.Sprintf("videos/%s/%d-%s", uid, time.Now().Unix(), req.FileName)
	log.Printf("InitiateUpload: using bucket '%s', object '%s'", cfg.GcsBucket, objectName)
	
	if cfg.GcsBucket == "" {
		log.Printf("InitiateUpload: ERROR - GCS bucket name is empty!")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GCS bucket not configured"})
		return
	}

	// Create a signed URL for PUT request
	url, err := storageClient.Bucket(cfg.GcsBucket).SignedURL(objectName, &storage.SignedURLOptions{
		Method:      "PUT",
		Expires:     time.Now().Add(15 * time.Minute),
		ContentType: req.FileType,
	})
	if err != nil {
		log.Printf("Failed to generate signed URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate upload"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"uploadUrl":  url,
		"objectName": objectName,
	})
}

// FinalizeUpload creates the video record after the file is in GCS.
func FinalizeUpload(c *gin.Context) {
	uid := c.GetString("uid")
	var req struct {
		ObjectName  string `json:"objectName" binding:"required"`
		Title       string `json:"title" binding:"required"`
		Description string `json:"description" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "objectName, title, and description are required"})
		return
	}

	// Generate thumbnail from the uploaded video
	thumbnailURL, err := generateThumbnail(req.ObjectName, uid)
	if err != nil {
		log.Printf("FinalizeUpload: thumbnail generation failed: %v", err)
		// Continue without thumbnail rather than failing the entire upload
		thumbnailURL = ""
	}

	vid := models.Video{
		ID:           uuid.NewString(),
		UserID:       uid,
		Title:        req.Title,
		Description:  req.Description,
		ObjectName:   req.ObjectName,
		ThumbnailURL: thumbnailURL,
	}
	if err := db.Conn.Create(&vid).Error; err != nil {
		log.Printf("FinalizeUpload: db.Conn.Create error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	go ai.GenerateAndCacheSummary(vid.ID, "gs://"+cfg.GcsBucket+"/"+req.ObjectName)

	c.JSON(http.StatusCreated, vid)
}

// generateThumbnail downloads video from GCS, generates thumbnail, uploads it back
func generateThumbnail(objectName, uid string) (string, error) {
	ctx := context.Background()
	
	// Create temporary files
	tempVideo, err := os.CreateTemp("", "video-*.mp4")
	if err != nil {
		return "", fmt.Errorf("failed to create temp video file: %v", err)
	}
	defer os.Remove(tempVideo.Name())
	defer tempVideo.Close()

	// Download video from GCS
	reader, err := storageClient.Bucket(cfg.GcsBucket).Object(objectName).NewReader(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create GCS reader: %v", err)
	}
	defer reader.Close()

	if _, err := io.Copy(tempVideo, reader); err != nil {
		return "", fmt.Errorf("failed to download video: %v", err)
	}
	tempVideo.Close()

	// Get video duration for thumbnail timing
	metadata, err := fluentffmpeg.Probe(tempVideo.Name())
	if err != nil {
		return "", fmt.Errorf("failed to probe video: %v", err)
	}

	formatData, ok := metadata["format"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("format data not found in metadata")
	}

	durationStr, ok := formatData["duration"].(string)
	if !ok {
		return "", fmt.Errorf("duration not found in format data")
	}

	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return "", fmt.Errorf("failed to parse duration: %v", err)
	}

	// Generate thumbnail at 1/4 of video duration
	seekTime := duration / 4
	hours := int(seekTime / 3600)
	minutes := int((seekTime - float64(hours*3600)) / 60)
	seconds := int(seekTime - float64(hours*3600) - float64(minutes*60))
	seekTimeString := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)

	// Generate thumbnail
	buf := bytes.NewBuffer(nil)
	err = fluentffmpeg.NewCommand("").
		InputPath(tempVideo.Name()).
		OutputFormat("image2").
		OutputOptions("-vframes", "1", "-ss", seekTimeString).
		PipeOutput(buf).Run()
	if err != nil {
		return "", fmt.Errorf("ffmpeg thumbnail generation failed: %v", err)
	}

	// Upload thumbnail to GCS
	thumbnailObject := fmt.Sprintf("thumbnails/%s/%d-thumbnail.jpg", uid, time.Now().Unix())
	thumbnailWriter := storageClient.Bucket(cfg.GcsBucket).Object(thumbnailObject).NewWriter(ctx)
	thumbnailWriter.ContentType = "image/jpeg"

	if _, err := io.Copy(thumbnailWriter, buf); err != nil {
		return "", fmt.Errorf("failed to upload thumbnail: %v", err)
	}

	if err := thumbnailWriter.Close(); err != nil {
		return "", fmt.Errorf("failed to close thumbnail writer: %v", err)
	}

	thumbnailURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", cfg.GcsBucket, thumbnailObject)
	return thumbnailURL, nil
}


func GetVideos(c *gin.Context) {
    var videos []models.Video
    err := db.Conn.
        Preload("User").
        Order("created_at ASC").
        Find(&videos).Error

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
        return
    }

    c.JSON(http.StatusOK, videos)
}

func GetVideo(c *gin.Context) {
	var video models.Video
	if err := db.Conn.Preload("User").First(&video, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
		return
	}

	// Get like count
	var likeCount int64
	db.Conn.Model(&models.Like{}).Where("video_id = ?", video.ID).Count(&likeCount)
	video.Likes = int(likeCount)

	// Check if user has liked the video
	uid, ok := c.Get("uid")
	if ok {
		var like models.Like
		if err := db.Conn.First(&like, "user_id = ? AND video_id = ?", uid, video.ID).Error; err == nil {
			video.IsLiked = true
		}
	}

	c.JSON(http.StatusOK, video)
}

func IncrementView(c *gin.Context) {
	var video models.Video
	if err := db.Conn.First(&video, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
		return
	}
	db.Conn.Model(&video).Update("views", gorm.Expr("views + 1"))
	c.Status(http.StatusOK)
}

func ToggleLike(c *gin.Context) {
	uid := c.GetString("uid")
	vid := c.Param("id")

	var like models.Like
	if err := db.Conn.First(&like, "user_id = ? AND video_id = ?", uid, vid).Error; err == nil {
		db.Conn.Delete(&like)
		c.Status(http.StatusOK)
		return
	}

	like = models.Like{UserID: uid, VideoID: vid}
	if err := db.Conn.Create(&like).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	c.Status(http.StatusOK)
}

// CreateLike adds a like to a video (idempotent - safe to call multiple times)
func CreateLike(c *gin.Context) {
	uid := c.GetString("uid")
	vid := c.Param("id")

	// Check if video exists
	var video models.Video
	if err := db.Conn.First(&video, "id = ?", vid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
		return
	}

	// Check if like already exists
	var existingLike models.Like
	if err := db.Conn.First(&existingLike, "user_id = ? AND video_id = ?", uid, vid).Error; err == nil {
		// Like already exists - this is fine (idempotent)
		c.Status(http.StatusOK)
		return
	}

	// Create new like
	like := models.Like{UserID: uid, VideoID: vid}
	if err := db.Conn.Create(&like).Error; err != nil {
		// Handle race condition where like was created between check and create
		if err := db.Conn.First(&existingLike, "user_id = ? AND video_id = ?", uid, vid).Error; err == nil {
			c.Status(http.StatusOK)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	c.Status(http.StatusOK)
}

// RemoveLike removes a like from a video (idempotent - safe to call multiple times)
func RemoveLike(c *gin.Context) {
	uid := c.GetString("uid")
	vid := c.Param("id")

	// Delete the like (if it exists)
	result := db.Conn.Where("user_id = ? AND video_id = ?", uid, vid).Delete(&models.Like{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// Success whether like existed or not (idempotent)
	c.Status(http.StatusOK)
}