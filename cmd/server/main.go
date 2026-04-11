package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rally-go/rally-realtime/internal/config"
	"github.com/rally-go/rally-realtime/internal/firebase"
	"github.com/rally-go/rally-realtime/internal/pubsub"
	"github.com/rally-go/rally-realtime/internal/socket"
	"github.com/rally-go/rally-realtime/internal/version"
)

func main() {
	cfg := config.Load()

	log.Printf("Starting Rally Realtime Server %s", version.Version)
	log.Printf("Commit SHA: %s, Build Time: %s", version.CommitSHA, version.BuildTime)

	// Initialise Firebase Auth
	firebase.MustInitialize(cfg.Firebase.CredentialsPath)

	// Initialise Redis pub/sub
	redisPubSub, err := pubsub.NewRedisPubSub(cfg.Redis.Addr)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisPubSub.Close()

	// Parse allowed origins (comma-separated)
	var allowedOrigins []string
	if cfg.Server.AllowedOrigins != "" {
		for _, o := range strings.Split(cfg.Server.AllowedOrigins, ",") {
			if trimmed := strings.TrimSpace(o); trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	// Initialise WebSocket hub and server
	hub := socket.NewHub(redisPubSub)
	go hub.Run()

	wsServer := socket.NewServer(hub, firebase.GetAuthClient(), allowedOrigins)

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
	mux.HandleFunc("/ws", wsServer.ServeWs)

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server listening on port %s", cfg.Server.Port)
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
