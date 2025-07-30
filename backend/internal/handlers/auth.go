// This file contains the "handlers" for all authentication-related actions.
// It manages user registration, checking for existing usernames, and fetching
// user profiles. It works closely with Firebase to ensure users are who they say they are.
package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/hi-wesley/mini-youtube/internal/db"
	"github.com/hi-wesley/mini-youtube/internal/firebase"
	"github.com/hi-wesley/mini-youtube/internal/models"
)

// POST /v1/auth/check-username  {username}
func CheckUsername(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User
	if err := db.Conn.Where("LOWER(username) = LOWER(?)", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
		return
	}

	c.Status(http.StatusOK)
}

// POST /v1/auth/register  {username} with Authorization header
func RegisterUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User
	if err := db.Conn.Where("LOWER(username) = LOWER(?)", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
		return
	}

	// Get token from Authorization header
	h := c.GetHeader("Authorization")
	idToken := strings.TrimPrefix(h, "Bearer ")
	if idToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth token"})
		return
	}

	// Verify the Firebase token
	token, err := firebase.Client.VerifyIDToken(c, idToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
		return
	}

	uid := token.UID
	email, _ := token.Claims["email"].(string)

	user := models.User{ID: uid, Email: email, Username: req.Username}
	if err := db.Conn.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already taken"})
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
