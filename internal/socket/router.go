package socket

import (
	"log"
	"net/http"

	"firebase.google.com/go/v4/auth"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rally-go/rally-realtime/internal/middleware"
)

// Server holds the dependencies for the WebSocket HTTP handler.
type Server struct {
	hub          *Hub
	firebaseAuth *auth.Client
	upgrader     websocket.Upgrader
}

// NewServer creates a Server. allowedOrigins is the list of permitted Origin
// header values; if empty, all origins are allowed (suitable for development).
func NewServer(hub *Hub, firebaseAuth *auth.Client, allowedOrigins []string) *Server {
	allowedSet := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowedSet[o] = true
	}

	return &Server{
		hub:          hub,
		firebaseAuth: firebaseAuth,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				if len(allowedOrigins) == 0 {
					return true // development: allow all origins
				}
				return allowedSet[r.Header.Get("Origin")]
			},
		},
	}
}

// ServeWs handles WebSocket upgrade requests.
// The client must supply:
//   - room_id  — the room to join
//   - token    — a valid Firebase ID token (user_id is derived from the token)
func (s *Server) ServeWs(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room_id")
	if roomID == "" {
		http.Error(w, "room_id is required", http.StatusBadRequest)
		return
	}

	// Authenticate before upgrading; browsers cannot send auth headers for WS.
	userID, err := middleware.VerifyWSToken(s.firebaseAuth, r)
	if err != nil {
		log.Printf("WebSocket auth failed: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	clientID := uuid.New().String()
	client := NewClient(clientID, userID, roomID, s.hub, conn)

	s.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()

	log.Printf("New WebSocket connection: client=%s user=%s room=%s", clientID, userID, roomID)
}
