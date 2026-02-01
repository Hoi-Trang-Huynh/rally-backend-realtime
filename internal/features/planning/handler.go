package planning

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// PlanningAction represents a planning action payload.
type PlanningAction struct {
	Action    string    `json:"action"` // "lock", "unlock", "update"
	ItemID    string    `json:"item_id"`
	UserID    string    `json:"user_id"`
	Data      any       `json:"data,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ItemLock represents a lock on an itinerary item.
type ItemLock struct {
	UserID    string
	ExpiresAt time.Time
}

// Handler handles planning/collaboration operations.
type Handler struct {
	// In-memory locks (TODO: Use Redis for distributed locking)
	locks map[string]*ItemLock
	mu    sync.RWMutex
}

// NewHandler creates a new planning handler.
func NewHandler() *Handler {
	h := &Handler{
		locks: make(map[string]*ItemLock),
	}

	// Start lock cleanup goroutine
	go h.cleanupExpiredLocks()

	return h
}

// ProcessAction processes an incoming planning action.
func (h *Handler) ProcessAction(userID string, payload json.RawMessage) (*PlanningAction, error) {
	var action PlanningAction
	if err := json.Unmarshal(payload, &action); err != nil {
		return nil, err
	}

	action.UserID = userID
	action.Timestamp = time.Now()

	switch action.Action {
	case "lock":
		if err := h.lockItem(action.ItemID, userID); err != nil {
			log.Printf("Failed to lock item %s: %v", action.ItemID, err)
			return nil, err
		}
	case "unlock":
		h.unlockItem(action.ItemID, userID)
	case "update":
		// TODO: Validate lock ownership before allowing update
		log.Printf("Planning update: item=%s user=%s", action.ItemID, userID)
	}

	return &action, nil
}

// lockItem attempts to acquire a lock on an item.
func (h *Handler) lockItem(itemID, userID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if already locked by another user
	if lock, exists := h.locks[itemID]; exists {
		if lock.UserID != userID && time.Now().Before(lock.ExpiresAt) {
			log.Printf("Item %s is locked by user %s", itemID, lock.UserID)
			return nil // Item is locked
		}
	}

	// Acquire lock (expires in 5 minutes)
	h.locks[itemID] = &ItemLock{
		UserID:    userID,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	log.Printf("Item %s locked by user %s", itemID, userID)
	return nil
}

// unlockItem releases a lock on an item.
func (h *Handler) unlockItem(itemID, userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if lock, exists := h.locks[itemID]; exists && lock.UserID == userID {
		delete(h.locks, itemID)
		log.Printf("Item %s unlocked by user %s", itemID, userID)
	}
}

// cleanupExpiredLocks periodically removes expired locks.
func (h *Handler) cleanupExpiredLocks() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		h.mu.Lock()
		now := time.Now()
		for itemID, lock := range h.locks {
			if now.After(lock.ExpiresAt) {
				delete(h.locks, itemID)
				log.Printf("Lock expired: item=%s user=%s", itemID, lock.UserID)
			}
		}
		h.mu.Unlock()
	}
}
