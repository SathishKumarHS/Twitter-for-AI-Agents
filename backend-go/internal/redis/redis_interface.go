// Package redis provides Redis caching for feeds, agent status, and distributed turn locking.
package redis

import (
	"context"
	"time"
)

// Storage and locking contract for Redis caching and concurrency control
type RedisInterface interface {
	// Retrieves the JSON string payload for a cached timeline feed limit, or nil if miss
	GetFeed(ctx context.Context, limit int) *string
	// Caches a timeline feed JSON payload for a specified limit
	SetFeed(ctx context.Context, limit int, payload []byte) error
	// Removes cached timeline feed keys when new posts are created
	InvalidateFeed(ctx context.Context) error

	// Checks if an agent's active status is cached in Redis
	IsAgentActive(ctx context.Context, id int64) (active bool, found bool)
	// Caches an agent's active status in Redis
	SetAgentActive(ctx context.Context, id int64, active bool) error

	// Tries to acquire a distributed SetNX turn lock for an agent ID with specified TTL
	Acquire(ctx context.Context, id int64, ttl time.Duration) (func(), bool, error)

	// Closes the underlying Redis connection client
	Close() error
}
