package socket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/rally-go/rally-realtime/internal/pubsub"
)

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	// Registered clients by room
	Rooms map[string]map[*Client]bool

	// All registered clients
	Clients map[*Client]bool

	// Inbound messages from clients
	Broadcast chan *BroadcastMessage

	// Register requests from clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Redis pub/sub for cross-server communication
	PubSub pubsub.PubSub

	// Mutex for thread-safe access
	mu sync.RWMutex
}

// BroadcastMessage represents a message to be broadcast.
type BroadcastMessage struct {
	RoomID  string
	Message []byte
	Sender  *Client // nil if from Redis
}

// NewHub creates a new Hub instance.
func NewHub(pubsub pubsub.PubSub) *Hub {
	return &Hub{
		Rooms:      make(map[string]map[*Client]bool),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan *BroadcastMessage, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		PubSub:     pubsub,
	}
}

// Run starts the hub's main loop.
func (h *Hub) Run() {
	// Subscribe to Redis messages
	if h.PubSub != nil {
		go h.subscribeToRedis()
	}

	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.Broadcast:
			h.broadcastToRoom(message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.Clients[client] = true

	if _, ok := h.Rooms[client.RoomID]; !ok {
		h.Rooms[client.RoomID] = make(map[*Client]bool)
	}
	h.Rooms[client.RoomID][client] = true

	log.Printf("Client %s joined room %s (total in room: %d)",
		client.ID, client.RoomID, len(h.Rooms[client.RoomID]))
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.Clients[client]; ok {
		delete(h.Clients, client)
		close(client.Send)

		if room, ok := h.Rooms[client.RoomID]; ok {
			delete(room, client)
			if len(room) == 0 {
				delete(h.Rooms, client.RoomID)
			}
		}

		log.Printf("Client %s left room %s", client.ID, client.RoomID)
	}
}

func (h *Hub) broadcastToRoom(msg *BroadcastMessage) {
	h.mu.RLock()
	clients, ok := h.Rooms[msg.RoomID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	for client := range clients {
		// Don't send back to sender (unless from Redis)
		if msg.Sender != nil && client == msg.Sender {
			continue
		}

		select {
		case client.Send <- msg.Message:
		default:
			// Client's send buffer is full, close connection
			h.mu.Lock()
			close(client.Send)
			delete(h.Clients, client)
			delete(h.Rooms[msg.RoomID], client)
			h.mu.Unlock()
		}
	}
}

// RouteMessage routes incoming messages to appropriate handlers.
func (h *Hub) RouteMessage(client *Client, msg *Message) {
	// Create outbound message
	outbound, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	// Broadcast to local clients
	h.Broadcast <- &BroadcastMessage{
		RoomID:  msg.RoomID,
		Message: outbound,
		Sender:  client,
	}

	// Publish to Redis for other server instances
	if h.PubSub != nil {
		channel := "room:" + msg.RoomID
		if err := h.PubSub.Publish(channel, outbound); err != nil {
			log.Printf("Failed to publish to Redis: %v", err)
		}
	}

	// Route to specific feature handler based on message type
	switch msg.Type {
	case "chat":
		h.handleChat(client, msg)
	case "location":
		h.handleLocation(client, msg)
	case "planning":
		h.handlePlanning(client, msg)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

func (h *Hub) subscribeToRedis() {
	// Subscribe to all room messages using pattern
	messages := h.PubSub.Subscribe("room:*")

	for msg := range messages {
		// Extract room ID from channel name
		roomID := msg.Channel[5:] // Remove "room:" prefix

		h.Broadcast <- &BroadcastMessage{
			RoomID:  roomID,
			Message: msg.Payload,
			Sender:  nil, // From Redis, not a local client
		}
	}
}

// Feature handlers (to be expanded in features package)
func (h *Hub) handleChat(client *Client, msg *Message) {
	log.Printf("Chat message from %s in room %s", client.UserID, msg.RoomID)
	// TODO: Save to MongoDB, sanitize text
}

func (h *Hub) handleLocation(client *Client, msg *Message) {
	log.Printf("Location update from %s in room %s", client.UserID, msg.RoomID)
	// TODO: Update Firestore, filter coordinates
}

func (h *Hub) handlePlanning(client *Client, msg *Message) {
	log.Printf("Planning update from %s in room %s", client.UserID, msg.RoomID)
	// TODO: Lock itinerary item for collaboration
}
