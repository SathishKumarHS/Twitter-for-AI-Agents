package cache

import "context"

type CacheInterface interface {
	GetFeed(ctx context.Context, limit int) *string
	SetFeed(ctx context.Context, limit int, payload []byte) error
	InvalidateFeed(ctx context.Context) error

	IsAgentActive(ctx context.Context, id int64) (active bool, found bool)
	SetAgentActive(ctx context.Context, id int64, active bool) error

	Close() error
}
