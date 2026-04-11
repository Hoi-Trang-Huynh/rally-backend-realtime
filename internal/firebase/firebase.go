package firebase

import (
	"context"
	"log"
	"sync"

	fb "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var (
	app        *fb.App
	authClient *auth.Client
	once       sync.Once
)

// InitializeClient initialises the Firebase app and Auth client once.
// Pass an empty credentialsPath on Cloud Run to use Application Default Credentials.
func InitializeClient(credentialsPath string) error {
	var err error
	once.Do(func() {
		ctx := context.Background()

		if credentialsPath != "" {
			app, err = fb.NewApp(ctx, nil, option.WithCredentialsFile(credentialsPath))
		} else {
			app, err = fb.NewApp(ctx, nil)
		}
		if err != nil {
			log.Printf("Error initialising Firebase app: %v", err)
			return
		}

		authClient, err = app.Auth(ctx)
		if err != nil {
			log.Printf("Error initialising Firebase Auth client: %v", err)
			return
		}

		log.Println("Firebase initialised successfully")
	})
	return err
}

// GetAuthClient returns the singleton Firebase Auth client.
// Panics if InitializeClient has not been called successfully first.
func GetAuthClient() *auth.Client {
	if authClient == nil {
		log.Fatal("Firebase auth client not initialised — call InitializeClient() first")
	}
	return authClient
}

// MustInitialize calls InitializeClient and fatals on error.
func MustInitialize(credentialsPath string) {
	if err := InitializeClient(credentialsPath); err != nil {
		log.Fatalf("Failed to initialise Firebase: %v", err)
	}
}
