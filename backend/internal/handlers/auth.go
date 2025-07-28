package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hi-wesley/mini-youtube/internal/db"
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
	if err := db.Conn.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
		return
	}

	c.Status(http.StatusOK)
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

	var existingUser models.User
    if err := db.Conn.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
        return
    }

	uid := c.GetString("uid")
	email := c.GetString("email")

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
