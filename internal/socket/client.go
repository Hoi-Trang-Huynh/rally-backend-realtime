package socket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 4096
)

// Client represents a single WebSocket connection.
type Client struct {
	ID     string
	UserID string
	RoomID string
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
}

// MessageType represents the type of a WebSocket message.
type MessageType string

const (
	MessageTypeChat     MessageType = "chat"
	MessageTypeLocation MessageType = "location"
	MessageTypePlanning MessageType = "planning"
)

// IsValid checks if the message type is supported.
func (t MessageType) IsValid() bool {
	switch t {
	case MessageTypeChat, MessageTypeLocation, MessageTypePlanning:
		return true
	}
	return false
}

// Message represents a WebSocket message structure.
type Message struct {
	Type    MessageType     `json:"type"`
	RoomID  string          `json:"room_id"`
	Payload json.RawMessage `json:"payload"`
}

// NewClient creates a new client instance.
func NewClient(id, userID, roomID string, hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		ID:     id,
		UserID: userID,
		RoomID: roomID,
		Hub:    hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse and route the message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Invalid message format: %v", err)
			continue
		}

		// Validate message type
		if !msg.Type.IsValid() {
			log.Printf("Unsupported message type: %s", msg.Type)
			continue
		}

		// Set room ID from client if not in message
		if msg.RoomID == "" {
			msg.RoomID = c.RoomID
		}

		c.Hub.RouteMessage(c, &msg)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
