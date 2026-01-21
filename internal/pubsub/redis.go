package pubsub

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

// RedisPubSub implements PubSub using Redis.
type RedisPubSub struct {
	client *redis.Client
	ctx    context.Context
	cancel context.CancelFunc
}

// NewRedisPubSub creates a new Redis pub/sub instance.
func NewRedisPubSub(addr string) (*RedisPubSub, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // No password by default
		DB:       0,  // Default DB
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		cancel()
		return nil, err
	}

	log.Printf("Connected to Redis at %s", addr)

	return &RedisPubSub{
		client: client,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Publish sends a message to a channel.
func (r *RedisPubSub) Publish(channel string, message []byte) error {
	return r.client.Publish(r.ctx, channel, message).Err()
}

// Subscribe returns a channel that receives messages from matching channels.
func (r *RedisPubSub) Subscribe(pattern string) <-chan PubSubMessage {
	messages := make(chan PubSubMessage, 256)

	go func() {
		defer close(messages)

		// Use PSubscribe for pattern matching
		pubsub := r.client.PSubscribe(r.ctx, pattern)
		defer pubsub.Close()

		ch := pubsub.Channel()

		for {
			select {
			case <-r.ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				messages <- PubSubMessage{
					Channel: msg.Channel,
					Payload: []byte(msg.Payload),
				}
			}
		}
	}()

	return messages
}

// Close closes the Redis connection.
func (r *RedisPubSub) Close() error {
	r.cancel()
	return r.client.Close()
}
