package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/<org>/mini-youtube/internal/config"
	"github.com/<org>/mini-youtube/internal/db"
	"github.com/<org>/mini-youtube/internal/handlers"
	"github.com/<org>/mini-youtube/internal/middleware"
)

func main() {
	// ----- configuration & database -----
	cfg := config.Load()
	if err := db.Connect(cfg.DB); err != nil {
		log.Fatalf("db connect: %v", err)
	}
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// ----- HTTP router -----
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(p gin.LogFormatterParams) string {
			return fmt.Sprintf(`{"time":"%s","method":"%s","path":"%s","status":%d,"latency":"%s"}`+"\n",
				p.TimeStamp.Format(time.RFC3339), p.Method, p.Path, p.StatusCode, p.Latency)
		},
		Output: os.Stdout,
	}))

	// public
	v1 := router.Group("/v1")
	v1.POST("/auth/register", handlers.RegisterUser)
	v1.POST("/auth/login", handlers.LoginUser)
	v1.GET("/videos/:id", handlers.GetVideo)

	// auth‑protected
	auth := v1.Group("/")
	auth.Use(middleware.Auth())
	{
		auth.GET("/profile", handlers.GetProfile)
		auth.POST("/videos", handlers.UploadVideo)
		auth.POST("/videos/:id/like", handlers.ToggleLike)
		auth.GET("/ws/comments", handlers.CommentsSocket) // ws://…/ws/comments?vid=<id>
		auth.POST("/comments", handlers.CreateComment)
	}

	// health‑check
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
