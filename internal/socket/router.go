package socket

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking for production
		return true
	},
}

// ServeWs handles WebSocket requests from clients.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Extract parameters from query string
	roomID := r.URL.Query().Get("room_id")
	userID := r.URL.Query().Get("user_id")

	if roomID == "" {
		http.Error(w, "room_id is required", http.StatusBadRequest)
		return
	}

	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create new client
	clientID := uuid.New().String()
	client := NewClient(clientID, userID, roomID, hub, conn)

	// Register client with hub
	hub.Register <- client

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()

	log.Printf("New WebSocket connection: client=%s user=%s room=%s", clientID, userID, roomID)
}
