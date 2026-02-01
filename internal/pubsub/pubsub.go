package pubsub

// PubSubMessage represents a message received from pub/sub.
type PubSubMessage struct {
	Channel string
	Payload []byte
}

// PubSub defines the interface for pub/sub operations.
type PubSub interface {
	// Publish sends a message to a channel.
	Publish(channel string, message []byte) error

	// Subscribe returns a channel that receives messages from matching channels.
	// Pattern supports Redis pattern matching (e.g., "room:*").
	Subscribe(pattern string) <-chan PubSubMessage

	// Close closes the pub/sub connection.
	Close() error
}
