package lock

import (
	"context"
	"testing"
	"time"
)

func TestRedisLockerConcurrency(t *testing.T) {
	ctx := context.Background()
	r := NewRedis("localhost:16379")
	defer r.Close()

	if _, err := r.client.Ping(ctx).Result(); err != nil {
		t.Skipf("redis not available: %v", err)
	}

	// First holder acquires; second must be rejected while held.
	release1, ok, err := r.Acquire(ctx, 42, 30*time.Second)
	if err != nil || !ok {
		t.Fatalf("first acquire: ok=%v err=%v", ok, err)
	}
	_, ok2, err := r.Acquire(ctx, 42, 30*time.Second)
	if err != nil || ok2 {
		t.Fatalf("second acquire should fail while held: ok=%v err=%v", ok2, err)
	}

	// Different agent id is independent.
	release3, ok3, err := r.Acquire(ctx, 43, 30*time.Second)
	if err != nil || !ok3 {
		t.Fatalf("different agent should acquire: ok=%v err=%v", ok3, err)
	}
	release3()

	// Release -> lock becomes available again.
	release1()
	release, ok4, err := r.Acquire(ctx, 42, 30*time.Second)
	if err != nil || !ok4 {
		t.Fatalf("re-acquire after release: ok=%v err=%v", ok4, err)
	}
	release()

	// TTL expiry frees a lock whose holder died without releasing.
	held, _, _ := r.Acquire(ctx, 99, 1*time.Second)
	held()
	time.Sleep(1500 * time.Millisecond)
	_, ok5, err := r.Acquire(ctx, 99, 30*time.Second)
	if err != nil || !ok5 {
		t.Fatalf("expired lock should be re-acquirable: ok=%v err=%v", ok5, err)
	}
}
