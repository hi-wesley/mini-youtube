package handlers

import (
	"context"
	"net/http"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"

	"github.com/hi-wesley/mini-youtube/internal/db"
	"github.com/hi-wesley/mini-youtube/internal/models"
)

var (
	fbApp    *firebase.App
	fbClient *auth.Client
)

func init() {
	var err error
	fbApp, err = firebase.NewApp(context.Background(), nil, option.WithCredentialsFile(
		// GOOGLE_APPLICATION_CREDENTIALS env var already points to the JSON
		"",
	))
	if err != nil { panic(err) }

	fbClient, err = fbApp.Auth(context.Background())
	if err != nil { panic(err) }
}

// POST /v1/auth/register  {username}
func RegisterUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid := c.GetString("uid")
	email := c.GetString("email")

	user := models.User{ID: uid, Email: email, Username: req.Username}
	if err := db.Conn.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username or email exists"})
		return
	}
	c.Status(http.StatusCreated)
}

// POST /v1/auth/login  {email, password}
// Firebase handles password check client‑side (web SDK). This endpoint exists only
// if you need a custom token; otherwise remove it.
func LoginUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "use Firebase client SDK to login"})
}

// GET /v1/profile
func GetProfile(c *gin.Context) {
	uid := c.GetString("uid")
	var u models.User
	if err := db.Conn.First(&u, "id = ?", uid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, u)
}
