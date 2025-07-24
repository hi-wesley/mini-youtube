package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hi-wesley/mini-youtube/internal/ai"
	"github.com/hi-wesley/mini-youtube/internal/config"
	"github.com/hi-wesley/mini-youtube/internal/db"
	"github.com/hi-wesley/mini-youtube/internal/models"
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

	log.Printf("UploadVideo: using GCS bucket: %s", cfg.GcsBucket)
	object := fmt.Sprintf("videos/%s/%d-%s", uid, time.Now().Unix(), header.Filename)
	log.Printf("UploadVideo: GCS object name: %s", object)

	writer := storageClient.Bucket(cfg.GcsBucket).Object(object).NewWriter(c)
	log.Println("UploadVideo: GCS writer created")

	if _, err := io.Copy(writer, file); err != nil {
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

	vid := models.Video{
		ID:          uuid.NewString(),
		UserID:      uid,
		Title:       c.PostForm("title"),
		Description: c.PostForm("description"),
		ObjectName:  object,
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

func GetVideo(c *gin.Context) {
	var video models.Video
	if err := db.Conn.First(&video, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
		return
	}
	c.JSON(http.StatusOK, video)
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