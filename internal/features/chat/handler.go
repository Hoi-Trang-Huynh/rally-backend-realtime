package chat

import (
	"encoding/json"
	"log"
	"time"
)

// ChatMessage represents a chat message payload.
type ChatMessage struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Handler handles chat-related operations.
type Handler struct {
	// TODO: Add MongoDB client for persistence
}

// NewHandler creates a new chat handler.
func NewHandler() *Handler {
	return &Handler{}
}

// ProcessMessage processes an incoming chat message.
func (h *Handler) ProcessMessage(userID string, payload json.RawMessage) (*ChatMessage, error) {
	var msg ChatMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		return nil, err
	}

	// Set metadata
	msg.UserID = userID
	msg.Timestamp = time.Now()

	// Sanitize content
	msg.Content = sanitizeText(msg.Content)

	// TODO: Save to MongoDB

	log.Printf("Chat message processed: user=%s content=%s", userID, msg.Content)

	return &msg, nil
}

// sanitizeText performs basic text sanitization.
func sanitizeText(text string) string {
	// TODO: Implement proper sanitization (XSS prevention, etc.)
	if len(text) > 2000 {
		text = text[:2000]
	}
	return text
}
