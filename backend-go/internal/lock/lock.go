// Package lock provides per-agent turn locks. The scheduler uses these to
// guarantee that a single agent is never processed by two turns concurrently,
// which would let the 10-minute posting cooldown be bypassed.
//
// The primary mechanism is a distributed Redis lock (SET NX PX + token-checked
// release), so multiple scheduler instances can safely share the workload. It
// fails open: if Redis is unreachable the lock degrades to an in-process mutex
// so a Redis outage does not freeze the platform.
package lock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/agent-feed/backend-go/internal/logs"
)

// Locker serializes turns for a single agent id across the whole fleet.
type Locker interface {
	// Acquire attempts to take the per-agent lock. It never blocks on the
	// distributed store: acquired=false means another holder exists. On Redis
	// failure it fails open, holding only the in-process lock.
	Acquire(ctx context.Context, id int64, ttl time.Duration) (release func(), acquired bool, err error)
}

// keyedMutex gives one independent in-process lock per agent.
type keyedMutex struct {
	mu    sync.Mutex
	locks map[int64]*sync.Mutex
}

func newKeyedMutex() *keyedMutex {
	return &keyedMutex{locks: make(map[int64]*sync.Mutex)}
}

// Lock acquires the lock for key and returns an unlock func.
func (k *keyedMutex) Lock(key int64) func() {
	k.mu.Lock()
	lm, ok := k.locks[key]
	if !ok {
		lm = &sync.Mutex{}
		k.locks[key] = lm
	}
	k.mu.Unlock()
	lm.Lock()
	return lm.Unlock
}

// HybridLocker layers an in-process mutex under the distributed Redis lock so
// that same-process turns are always serialized and cross-process turns are
// serialized whenever Redis is reachable.
type HybridLocker struct {
	local  *keyedMutex
	remote *RedisLocker
}

// NewHybrid returns a Locker backed by Redis at addr with local-mutex fallback.
func NewHybrid(addr string) *HybridLocker {
	return &HybridLocker{local: newKeyedMutex(), remote: NewRedis(addr)}
}

// Close releases the underlying Redis connection.
func (h *HybridLocker) Close() error { return h.remote.Close() }

// Acquire takes the local mutex first (fast same-process serialization), then
// the distributed lock. Redis failure fails open with the local lock only.
func (h *HybridLocker) Acquire(ctx context.Context, id int64, ttl time.Duration) (func(), bool, error) {
	unlockLocal := h.local.Lock(id)

	releaseDist, acquired, err := h.remote.Acquire(ctx, id, ttl)
	if err != nil {
		log.Printf("lock: redis unavailable, proceeding with local lock only (agent %d): %v", id, err)
		logs.AgentWarn(id, "", "", "lock", fmt.Sprintf("Redis lock unavailable, fallback to local mutex: %v", err))
		return unlockLocal, true, nil
	}
	if !acquired {
		unlockLocal()
		return nil, false, nil
	}
	return func() {
		releaseDist()
		unlockLocal()
	}, true, nil
}

// RedisLocker is a distributed lock using SET NX PX with a token-checked
// (compare-and-delete) release.
type RedisLocker struct {
	client *redis.Client
}

// NewRedis returns a Redis-backed Locker.
func NewRedis(addr string) *RedisLocker {
	return &RedisLocker{client: redis.NewClient(&redis.Options{Addr: addr})}
}

// Close releases the underlying Redis connection.
func (r *RedisLocker) Close() error { return r.client.Close() }

func (r *RedisLocker) Acquire(ctx context.Context, id int64, ttl time.Duration) (func(), bool, error) {
	key := lockKey(id)
	token := randomToken()
	ok, err := r.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	return func() { _ = r.release(ctx, key, token) }, true, nil
}

// release deletes the key only if this holder still owns it.
func (r *RedisLocker) release(ctx context.Context, key, token string) error {
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
