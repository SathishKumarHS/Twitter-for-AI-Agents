//go:build integration

package integration

import (
	"context"
	"testing"
	"time"
)

func TestRedisFeedCacheSemantics(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	if err := rc.SetFeed(ctx, 50, []byte("[1,2,3]")); err != nil {
		t.Fatalf("set feed: %v", err)
	}
	got := rc.GetFeed(ctx, 50)
	if got == nil || *got != "[1,2,3]" {
		t.Fatalf("get feed = %v, want [1,2,3]", got)
	}

	if err := rc.InvalidateFeed(ctx); err != nil {
		t.Fatalf("invalidate feed: %v", err)
	}
	if rc.GetFeed(ctx, 50) != nil {
		t.Error("feed not invalidated")
	}
}

func TestRedisAgentActiveCache(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	if err := rc.SetAgentActive(ctx, 7, true); err != nil {
		t.Fatalf("set active: %v", err)
	}
	active, found := rc.IsAgentActive(ctx, 7)
	if !found || !active {
		t.Fatalf("is active = %v/%v, want true/true", active, found)
	}

	if err := rc.SetAgentActive(ctx, 7, false); err != nil {
		t.Fatalf("set inactive: %v", err)
	}
	active, found = rc.IsAgentActive(ctx, 7)
	if !found || active {
		t.Fatalf("is active = %v/%v, want false/true", active, found)
	}
}

func TestRedisTurnLockExcludesAndReleases(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	release, acquired, err := rc.Acquire(ctx, 42, 10*time.Second)
	if err != nil || !acquired {
		t.Fatalf("first acquire: acquired=%v err=%v", acquired, err)
	}
	if raw.Exists(ctx, "agent:turn:42").Val() == 0 {
		t.Error("lock key missing while held")
	}

	_, acquired2, err := rc.Acquire(ctx, 42, 10*time.Second)
	if err != nil {
		t.Fatalf("second acquire err: %v", err)
	}
	if acquired2 {
		t.Error("second acquire succeeded while lock held")
	}

	release()
	if raw.Exists(ctx, "agent:turn:42").Val() != 0 {
		t.Error("lock key not deleted after release")
	}

	_, acquired3, err := rc.Acquire(ctx, 42, 10*time.Second)
	if err != nil || !acquired3 {
		t.Fatalf("re-acquire after release: acquired=%v err=%v", acquired3, err)
	}
}
