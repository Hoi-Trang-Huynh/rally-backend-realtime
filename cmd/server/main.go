package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rally-go/rally-realtime/internal/pubsub"
	"github.com/rally-go/rally-realtime/internal/socket"
	"github.com/rally-go/rally-realtime/internal/version"
)

func main() {
	// Get configuration from environment
	port := getEnv("PORT", "8080")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")

	log.Printf("Starting Rally Realtime Server %s", version.Version)
	log.Printf("Commit SHA: %s, Build Time: %s", version.CommitSHA, version.BuildTime)

	// Initialize Redis pub/sub
	redisPubSub, err := pubsub.NewRedisPubSub(redisAddr)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisPubSub.Close()

	// Initialize WebSocket hub
	hub := socket.NewHub(redisPubSub)
	go hub.Run()

	// Setup HTTP routes
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	// Version endpoint
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(version.Info())
	})

	// WebSocket endpoint
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		socket.ServeWs(hub, w, r)
	})

	// Create server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server listening on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
