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
		origins = strings.Split(cfg.AllowedOrigins, ",")
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
		v1.POST("/auth/login", handlers.LoginUser)
		v1.POST("/auth/check-username", handlers.CheckUsername)
		v1.POST("/auth/register", handlers.RegisterUser)

		v1.GET("/videos", handlers.GetVideos)
		v1.GET("/videos/:id", middleware.MaybeAuth(), handlers.GetVideo)
		v1.POST("/videos/:id/view", handlers.IncrementView)
		v1.GET("/videos/:id/comments", handlers.GetComments)
		v1.GET("/ws/comments", handlers.CommentsSocket) // ws://…/ws/comments?vid=<id>

		// auth-protected
		v1.GET("/profile", middleware.Auth(), handlers.GetProfile)
		v1.POST("/videos/initiate-upload", middleware.Auth(), handlers.InitiateUpload)
		v1.POST("/videos/finalize-upload", middleware.Auth(), handlers.FinalizeUpload)
		v1.POST("/videos/:id/like", middleware.Auth(), handlers.ToggleLike)
		v1.POST("/comments", middleware.Auth(), handlers.CreateComment)

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

