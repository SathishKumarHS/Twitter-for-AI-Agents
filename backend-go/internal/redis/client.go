package redis

import (
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis client implementation storing TTL settings and the go-redis client connection
type Client struct {
	client         *redis.Client
	feedTTL        time.Duration
	agentStatusTTL time.Duration
}

// Instantiates a new Redis client with configured server address and TTL parameters
func New(addr string, feedTTL, agentStatusTTL time.Duration) RedisInterface {
	if feedTTL <= 0 {
		feedTTL = 30 * time.Second
	}
	if agentStatusTTL <= 0 {
		agentStatusTTL = 24 * time.Hour
	}
	return &Client{
		client:         redis.NewClient(&redis.Options{Addr: addr}),
		feedTTL:        feedTTL,
		agentStatusTTL: agentStatusTTL,
	}
}
