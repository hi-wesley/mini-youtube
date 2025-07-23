package middleware

import (
	"context"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil { panic(err) }
	client, err := app.Auth(context.Background())
	if err != nil { panic(err) }

	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		idToken := strings.TrimPrefix(h, "Bearer ")
		if idToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing auth token"})
			return
		}
		token, err := client.VerifyIDToken(c, idToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
			return
		}
		c.Set("uid", token.UID)
		c.Next()
	}
}
