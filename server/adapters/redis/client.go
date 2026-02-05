package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps Redis operations for pub/sub and caching
type Client struct {
	rdb *redis.Client
}

// NewClient creates a Redis client from URL
func NewClient(redisURL string) (*Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(opts)
	return &Client{rdb: rdb}, nil
}

// Ping checks the Redis connection
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Publish sends a message to a channel
func (c *Client) Publish(ctx context.Context, channel string, message []byte) error {
	return c.rdb.Publish(ctx, channel, message).Err()
}

// Subscribe returns a subscription to channels
func (c *Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return c.rdb.Subscribe(ctx, channels...)
}

// Set stores a value with optional TTL
func (c *Client) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

// Get retrieves a value by key
func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	return c.rdb.Get(ctx, key).Bytes()
}

// Del removes keys
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

// Underlying returns the underlying redis.Client for advanced use
func (c *Client) Underlying() *redis.Client {
	return c.rdb
}
