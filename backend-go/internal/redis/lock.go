package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	agentLockKeyFmt = "agent:turn:%d"
)

// Tries to acquire a distributed SetNX turn lock for an agent ID; returns release unlock function and success flag
func (c *Client) Acquire(ctx context.Context, id int64, ttl time.Duration) (func(), bool, error) {
	if c == nil || c.client == nil {
		return func() {}, true, nil
	}

	key := lockKey(id)
	token := randomToken()

	acquired, err := c.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return nil, false, err
	}
	if !acquired {
		return nil, false, nil
	}

	return func() {
		relCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = c.release(relCtx, key, token)
	}, true, nil
}

// Releases a distributed lock using a Lua script to ensure atomic token comparison and deletion
func (c *Client) release(ctx context.Context, key, token string) error {
	const script = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
end
return 0
`
	return redis.NewScript(script).Run(ctx, c.client, []string{key}, token).Err()
}

// Formats the Redis lock key for an agent turn
func lockKey(id int64) string {
	return fmt.Sprintf(agentLockKeyFmt, id)
}

// Generates a cryptographically random token string for turn lock ownership
func randomToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
