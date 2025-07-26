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

func UploadVideo(c *gin.Context) {
	log.Println("UploadVideo: starting")
	uid := c.GetString("uid")
	log.Printf("UploadVideo: UID: %s", uid)

	file, header, err := c.Request.FormFile("video")
	if err != nil {
		log.Printf("UploadVideo: FormFile error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "video required"})
		return
	}
	defer file.Close()
	log.Println("UploadVideo: file retrieved from form")

	// Create a temporary file to store the video
	tempFile, err := os.CreateTemp("", "upload-*.mp4")
	if err != nil {
		log.Printf("UploadVideo: os.CreateTemp error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		return
	}
	defer os.Remove(tempFile.Name())

	// Copy the uploaded video to the temporary file
	if _, err := io.Copy(tempFile, file); err != nil {
		log.Printf("UploadVideo: io.Copy error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		return
	}

	log.Printf("UploadVideo: using GCS bucket: %s", cfg.GcsBucket)
	object := fmt.Sprintf("videos/%s/%d-%s", uid, time.Now().Unix(), header.Filename)
	log.Printf("UploadVideo: GCS object name: %s", object)

	writer := storageClient.Bucket(cfg.GcsBucket).Object(object).NewWriter(c)
	log.Println("UploadVideo: GCS writer created")

	// Reset the file pointer to the beginning of the file
	tempFile.Seek(0, 0)

	if _, err := io.Copy(writer, tempFile); err != nil {
		log.Printf("UploadVideo: io.Copy error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		return
	}
	log.Println("UploadVideo: file copied to GCS")

	if err := writer.Close(); err != nil {
		log.Printf("UploadVideo: writer.Close error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		return
	}
	log.Println("UploadVideo: GCS writer closed")

	

	// Get video duration
	metadata, err := fluentffmpeg.Probe(tempFile.Name())
	if err != nil {
		log.Printf("UploadVideo: ffmpeg probe error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get video metadata"})
		return
	}

	formatData, ok := metadata["format"].(map[string]interface{})
	if !ok {
		log.Printf("UploadVideo: format data not found in metadata")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get video duration"})
		return
	}

	durationStr, ok := formatData["duration"].(string)
	if !ok {
		log.Printf("UploadVideo: duration not found or not a string in format data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get video duration"})
		return
	}

	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		log.Printf("UploadVideo: failed to parse duration: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get video duration"})
		return
	}

	// Add strconv to imports
	// Add "strconv" to the import list
	// import (
	// 	"strconv"
	// )
	seekTime := duration / 4

	// Convert seekTime to HH:MM:SS format
	hours := int(seekTime / 3600)
	minutes := int((seekTime - float64(hours*3600)) / 60)
	seconds := int(seekTime - float64(hours*3600) - float64(minutes*60))
	seekTimeString := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)

	// Generate thumbnail
	buf := bytes.NewBuffer(nil)
	err = fluentffmpeg.NewCommand("").
		InputPath(tempFile.Name()).
		OutputFormat("image2").
		OutputOptions("-vframes", "1", "-ss", seekTimeString).
		PipeOutput(buf).Run()
	if err != nil {
		log.Printf("UploadVideo: ffmpeg error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "thumbnail generation failed"})
		return
	}

	thumbnailObject := fmt.Sprintf("thumbnails/%s/%d-thumbnail.jpg", uid, time.Now().Unix())
	thumbnailWriter := storageClient.Bucket(cfg.GcsBucket).Object(thumbnailObject).NewWriter(c)

	if _, err := io.Copy(thumbnailWriter, buf); err != nil {
		log.Printf("UploadVideo: thumbnail io.Copy error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "thumbnail upload failed"})
		return
	}

	if err := thumbnailWriter.Close(); err != nil {
		log.Printf("UploadVideo: thumbnail writer.Close error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "thumbnail upload failed"})
		return
	}

	

	vid := models.Video{
		ID:           uuid.NewString(),
		UserID:       uid,
		Title:        c.PostForm("title"),
		Description:  c.PostForm("description"),
		ObjectName:   object,
		ThumbnailURL: fmt.Sprintf("https://storage.googleapis.com/%s/%s", cfg.GcsBucket, thumbnailObject),
	}
	if err := db.Conn.Create(&vid).Error; err != nil {
		log.Printf("UploadVideo: db.Conn.Create error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	log.Println("UploadVideo: video record created in DB")

	go ai.GenerateAndCacheSummary(vid.ID, "gs://"+cfg.GcsBucket+"/"+object)
	log.Println("UploadVideo: async summary started")

	c.JSON(http.StatusCreated, vid)
	log.Println("UploadVideo: finished successfully")
}

func GetVideos(c *gin.Context) {
	var videos []models.Video
	if err := db.Conn.Find(&videos).Error; err != nil {
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
