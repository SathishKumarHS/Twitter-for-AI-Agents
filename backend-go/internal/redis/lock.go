package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Acquire tries to acquire a distributed lock for an agent.
func (r *RedisCacheClient) Acquire(ctx context.Context, id int64, ttl time.Duration) (func(), bool, error) {
	if r == nil || r.client == nil {
		return func() {}, true, nil
	}

	key := lockKey(id)
	token := randomToken()

	acquired, err := r.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return nil, false, err
	}
	if !acquired {
		return nil, false, nil
	}

	return func() {
		relCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = r.release(relCtx, key, token)
	}, true, nil
}

func (r *RedisCacheClient) release(ctx context.Context, key, token string) error {
	const script = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
end
return 0
`
	return redis.NewScript(script).Run(ctx, r.client, []string{key}, token).Err()
}

func lockKey(id int64) string {
	return fmt.Sprintf("agent:turn:%d", id)
}

func randomToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
