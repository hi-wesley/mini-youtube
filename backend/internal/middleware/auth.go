package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hi-wesley/mini-youtube/internal/firebase"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		log.Printf("Authorization header: %s", h)

		idToken := strings.TrimPrefix(h, "Bearer ")
		if idToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing auth token"})
			return
		}

		token, err := firebase.Client.VerifyIDToken(c, idToken)
		if err != nil {
			log.Printf("VerifyIDToken error: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
			return
		}

		c.Set("uid", token.UID)
		c.Set("email", token.Claims["email"])
		c.Next()
	}
}

// MaybeAuth is a middleware that will try to authenticate the user if a token is provided,
// but will not fail if the token is missing or invalid.
func MaybeAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		idToken := strings.TrimPrefix(h, "Bearer ")

		if idToken != "" {
			token, err := firebase.Client.VerifyIDToken(c, idToken)
			if err == nil && token != nil {
				c.Set("uid", token.UID)
				c.Set("email", token.Claims["email"])
			}
		}

		c.Next()
	}
}
