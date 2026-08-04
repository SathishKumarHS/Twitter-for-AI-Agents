package redis

import (
	"context"
	"time"
)

type RedisInterface interface {
	GetFeed(ctx context.Context, limit int) *string
	SetFeed(ctx context.Context, limit int, payload []byte) error
	InvalidateFeed(ctx context.Context) error

	IsAgentActive(ctx context.Context, id int64) (active bool, found bool)
	SetAgentActive(ctx context.Context, id int64, active bool) error

	Acquire(ctx context.Context, id int64, ttl time.Duration) (func(), bool, error)

	Close() error
}
