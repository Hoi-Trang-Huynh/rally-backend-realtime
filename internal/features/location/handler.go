package location

import (
	"encoding/json"
	"log"
	"time"
)

// LocationUpdate represents a location update payload.
type LocationUpdate struct {
	UserID    string    `json:"user_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Accuracy  float64   `json:"accuracy,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Handler handles location-related operations.
type Handler struct {
	// TODO: Add Firestore client for persistence
}

// NewHandler creates a new location handler.
func NewHandler() *Handler {
	return &Handler{}
}

// ProcessUpdate processes an incoming location update.
func (h *Handler) ProcessUpdate(userID string, payload json.RawMessage) (*LocationUpdate, error) {
	var loc LocationUpdate
	if err := json.Unmarshal(payload, &loc); err != nil {
		return nil, err
	}

	// Set metadata
	loc.UserID = userID
	loc.Timestamp = time.Now()

	// Validate and filter coordinates
	if !isValidCoordinate(loc.Latitude, loc.Longitude) {
		log.Printf("Invalid coordinates: lat=%f lng=%f", loc.Latitude, loc.Longitude)
		return nil, nil
	}

	// TODO: Update Firestore

	log.Printf("Location update processed: user=%s lat=%f lng=%f", userID, loc.Latitude, loc.Longitude)

	return &loc, nil
}

// isValidCoordinate checks if coordinates are within valid ranges.
func isValidCoordinate(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}
