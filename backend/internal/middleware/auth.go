package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	log.Println("Initializing Firebase...")
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase app: %v", err)
	}
	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize Firebase auth client: %v", err)
	}
	log.Println("Firebase initialized successfully")

	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		log.Printf("Authorization header: %s", h)

		idToken := strings.TrimPrefix(h, "Bearer ")
		if idToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing auth token"})
			return
		}

		token, err := client.VerifyIDToken(c, idToken)
		if err != nil {
			log.Printf("VerifyIDToken error: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
			return
		}

		c.Set("uid", token.UID)
		c.Next()
	}
}
