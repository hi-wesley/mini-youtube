package firebase

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
)

var Client *auth.Client

func init() {
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase app: %v", err)
	}
	Client, err = app.Auth(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize Firebase auth client: %v", err)
	}
	log.Println("Firebase initialized successfully")
}
