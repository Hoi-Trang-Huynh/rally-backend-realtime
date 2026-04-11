package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"firebase.google.com/go/v4/auth"
)

// VerifyWSToken extracts a Firebase ID token from the "token" query parameter,
// verifies it with Firebase Auth, and returns the verified Firebase UID on success.
//
// WebSocket connections from browsers cannot set custom HTTP headers, so the
// token is passed as a URL query parameter instead.
func VerifyWSToken(firebaseAuth *auth.Client, r *http.Request) (string, error) {
	token := r.URL.Query().Get("token")
	if token == "" {
		return "", fmt.Errorf("token query parameter is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t, err := firebaseAuth.VerifyIDToken(ctx, token)
	if err != nil {
		return "", fmt.Errorf("invalid or expired token: %w", err)
	}

	return t.UID, nil
}
