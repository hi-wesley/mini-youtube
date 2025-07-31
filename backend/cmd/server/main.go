// This file is the main entry point for the entire backend application.
// It starts up the web server, connects to the database, and defines all
// the API routes that the frontend can call. Think of it as the main switchboard
// that directs all incoming web traffic to the correct place.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/hi-wesley/mini-youtube/internal/config"
	"github.com/hi-wesley/mini-youtube/internal/db"
	"github.com/hi-wesley/mini-youtube/internal/handlers"
	"github.com/hi-wesley/mini-youtube/internal/middleware"
)

func main() {
	// ----- configuration & database -----
	cfg := config.Load()
	if err := db.Connect(cfg.DB); err != nil {
		log.Fatalf("db connect: %v", err)
	}
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("db automigrate: %v", err)
	}

	// ----- initialize rate limiter -----
	if cfg.RateLimitEnabled && cfg.RateLimitRedisURL != "" {
		if err := middleware.InitRateLimiter(cfg.RateLimitRedisURL, cfg.RateLimitRedisDB); err != nil {
			log.Printf("WARNING: Failed to initialize rate limiter: %v", err)
			log.Printf("Rate limiting will be disabled")
		} else {
			log.Printf("Rate limiting enabled with Redis")
		}
	} else {
		log.Printf("Rate limiting disabled")
	}

	// ----- HTTP router -----
	router := gin.New()
	router.RedirectTrailingSlash = true
	router.SetTrustedProxies(nil)
	router.Use(gin.Recovery())
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(p gin.LogFormatterParams) string {
			return fmt.Sprintf(`{"time":"%s","method":"%s","path":"%s","status":%d,"latency":"%s"}`+"\n",
				p.TimeStamp.Format(time.RFC3339), p.Method, p.Path, p.StatusCode, p.Latency)
		},
		Output: os.Stdout,
	}))

	// Set up CORS from environment variable
	origins := []string{"*"} // Default to allow all
	if cfg.AllowedOrigins != "" {
		// Handle both comma and space separated origins
		if strings.Contains(cfg.AllowedOrigins, ",") {
			origins = strings.Split(cfg.AllowedOrigins, ",")
		} else {
			origins = strings.Fields(cfg.AllowedOrigins) // Split by whitespace
		}
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
	}
	router.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// public
	v1 := router.Group("/v1")
	{
		// Authentication endpoints - daily limits
		v1.POST("/auth/login", middleware.RateLimitByIP(60, 24*time.Hour), handlers.LoginUser)
		v1.POST("/auth/check-username", middleware.RateLimitByIP(30, 24*time.Hour), handlers.CheckUsername)
		v1.POST("/auth/register", middleware.RateLimitByIP(6, 24*time.Hour), handlers.RegisterUser)

		// Public video endpoints - daily limits except comments
		v1.GET("/videos", middleware.RateLimitByIP(480, 24*time.Hour), handlers.GetVideos)
		v1.GET("/videos/:id", middleware.MaybeAuth(), middleware.RateLimitByIP(480, 24*time.Hour), handlers.GetVideo)
		v1.POST("/videos/:id/view", middleware.RateLimitByIP(480, 24*time.Hour), handlers.IncrementView)
		v1.GET("/videos/:id/comments", middleware.RateLimitByIP(60, time.Minute), handlers.GetComments)
		v1.GET("/ws/comments", handlers.CommentsSocket) // WebSocket - handled differently

		// auth-protected endpoints with user-based rate limiting
		v1.GET("/profile", middleware.Auth(), middleware.RateLimitByUser(60, time.Minute), handlers.GetProfile)
		v1.POST("/videos/initiate-upload", middleware.Auth(), middleware.RateLimitByUser(30, 24*time.Hour), handlers.InitiateUpload)
		v1.POST("/videos/finalize-upload", middleware.Auth(), middleware.RateLimitByUser(30, 24*time.Hour), handlers.FinalizeUpload)
		v1.POST("/videos/:id/like", middleware.Auth(), middleware.RateLimitByUser(60, 24*time.Hour), handlers.ToggleLike)
		v1.POST("/comments", middleware.Auth(), middleware.RateLimitByUser(30, 24*time.Hour), handlers.CreateComment)

	}

	// health
	router.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Printf("API listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("run: %v", err)
	}
}

