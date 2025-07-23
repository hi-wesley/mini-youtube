package handlers

import (
	"context"
	"net/http"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"

	"github.com/<org>/mini-youtube/internal/db"
	"github.com/<org>/mini-youtube/internal/models"
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

// POST /v1/auth/register  {email, password, username}
func RegisterUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Username string `json:"username" binding:"required,min=3"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// create in Firebase
	params := (&auth.UserToCreate{}).Email(req.Email).Password(req.Password)
	u, err := fbClient.CreateUser(c, params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// create in local DB
	user := models.User{ID: u.UID, Email: req.Email, Username: req.Username}
	if err := db.Conn.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username or email exists"})
		_ = fbClient.DeleteUser(c, u.UID) // roll back Firebase
		return
	}
	c.JSON(http.StatusCreated, gin.H{"uid": u.UID})
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
