package cache

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCacheClient struct {
	client         *redis.Client
	feedTTL        time.Duration
	agentStatusTTL time.Duration
}

// New returns a configured RedisCache instance.
func New(addr string, feedTTL, agentStatusTTL time.Duration) CacheInterface {
	if feedTTL <= 0 {
		feedTTL = 30 * time.Second
	}
	if agentStatusTTL <= 0 {
		agentStatusTTL = 24 * time.Hour
	}
	return &RedisCacheClient{
		client:         redis.NewClient(&redis.Options{Addr: addr}),
		feedTTL:        feedTTL,
		agentStatusTTL: agentStatusTTL,
	}
}