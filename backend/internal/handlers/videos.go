package handlers

import (
	"context"
	"fmt"
	"io"
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
		panic(err)
	}
	cfg = config.Load()
}

func UploadVideo(c *gin.Context) {
	uid := c.GetString("uid")
	file, header, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video required"})
		return
	}
	defer file.Close()

	// basic client‑side validated: size, mime type, etc.

	object := fmt.Sprintf("videos/%s/%d-%s", uid, time.Now().Unix(), header.Filename)
	writer := storageClient.Bucket(cfg.GcsBucket).Object(object).NewWriter(c)

	if _, err := io.Copy(writer, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		return
	}
	if err := writer.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		return
	}

	vid := models.Video{
		ID:          uuid.NewString(),
		UserID:      uid,
		Title:       c.PostForm("title"),
		Description: c.PostForm("description"),
		ObjectName:  object,
	}
	if err := db.Conn.Create(&vid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// Async summarization (fire‑and‑forget)
	go ai.GenerateAndCacheSummary(vid.ID, "gs://"+cfg.GcsBucket+"/"+object)

	c.JSON(http.StatusCreated, vid)
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