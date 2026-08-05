package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func setupTestRedis(t *testing.T) (RedisInterface, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	cli := New(mr.Addr(), 30*time.Second, 24*time.Hour)
	return cli, mr
}

func TestRedisNewDefaultTTL(t *testing.T) {
	clientInterface := New("localhost:6379", 0, 0)
	c, ok := clientInterface.(*Client)
	if !ok {
		t.Fatalf("expected *Client instance")
	}

	if c.feedTTL != 30*time.Second {
		t.Errorf("expected default feedTTL 30s, got %v", c.feedTTL)
	}
	if c.agentStatusTTL != 24*time.Hour {
		t.Errorf("expected default agentStatusTTL 24h, got %v", c.agentStatusTTL)
	}

	_ = clientInterface.Close()
}

func TestRedisNilClientSafety(t *testing.T) {
	var nilClient *Client

	ctx := context.Background()

	if err := nilClient.Close(); err != nil {
		t.Errorf("expected nil error on Close, got %v", err)
	}

	if feed := nilClient.GetFeed(ctx, 50); feed != nil {
		t.Errorf("expected nil feed on nil client, got %v", feed)
	}

	if err := nilClient.SetFeed(ctx, 50, []byte("data")); err != nil {
		t.Errorf("expected nil error on SetFeed, got %v", err)
	}

	if err := nilClient.InvalidateFeed(ctx); err != nil {
		t.Errorf("expected nil error on InvalidateFeed, got %v", err)
	}

	active, found := nilClient.IsAgentActive(ctx, 1)
	if active || found {
		t.Errorf("expected false, false for IsAgentActive on nil client, got %v, %v", active, found)
	}

	if err := nilClient.SetAgentActive(ctx, 1, true); err != nil {
		t.Errorf("expected nil error on SetAgentActive, got %v", err)
	}

	rel, acquired, err := nilClient.Acquire(ctx, 1, time.Second)
	if err != nil || !acquired || rel == nil {
		t.Errorf("expected fallback acquired=true on nil client lock")
	} else {
		rel()
	}
}

func TestRedisFeedCaching(t *testing.T) {
	cli, mr := setupTestRedis(t)
	defer mr.Close()
	defer cli.Close()

	ctx := context.Background()

	// Initial GetFeed should return nil
	feed := cli.GetFeed(ctx, 50)
	if feed != nil {
		t.Fatalf("expected nil feed before caching, got %s", *feed)
	}

	// Set feed cache
	payload := []byte(`[{"id": 1, "body": "hello"}]`)
	if err := cli.SetFeed(ctx, 50, payload); err != nil {
		t.Fatalf("failed to set feed cache: %v", err)
	}

	// GetFeed should return cached payload
	feed = cli.GetFeed(ctx, 50)
	if feed == nil || *feed != string(payload) {
		t.Fatalf("expected feed payload '%s', got '%v'", string(payload), feed)
	}

	// Invalidate feed cache
	if err := cli.InvalidateFeed(ctx); err != nil {
		t.Fatalf("failed to invalidate feed cache: %v", err)
	}

	feedAfter := cli.GetFeed(ctx, 50)
	if feedAfter != nil {
		t.Fatalf("expected nil feed after invalidation, got %s", *feedAfter)
	}
}

func TestRedisAgentActiveStatus(t *testing.T) {
	cli, mr := setupTestRedis(t)
	defer mr.Close()
	defer cli.Close()

	ctx := context.Background()
	agentID := int64(42)

	// Uncached agent should return found=false
	active, found := cli.IsAgentActive(ctx, agentID)
	if found || active {
		t.Fatalf("expected found=false, active=false for uncached agent")
	}

	// Set active = true
	if err := cli.SetAgentActive(ctx, agentID, true); err != nil {
		t.Fatalf("failed to set agent active: %v", err)
	}

	active, found = cli.IsAgentActive(ctx, agentID)
	if !found || !active {
		t.Fatalf("expected found=true, active=true, got found=%v, active=%v", found, active)
	}

	// Set active = false
	if err := cli.SetAgentActive(ctx, agentID, false); err != nil {
		t.Fatalf("failed to set agent inactive: %v", err)
	}

	active, found = cli.IsAgentActive(ctx, agentID)
	if !found || active {
		t.Fatalf("expected found=true, active=false, got found=%v, active=%v", found, active)
	}
}

func TestRedisDistributedLocking(t *testing.T) {
	cli, mr := setupTestRedis(t)
	defer mr.Close()
	defer cli.Close()

	ctx := context.Background()
	agentID := int64(100)
	ttl := 5 * time.Second

	// 1. First lock acquisition should succeed
	rel1, acquired1, err1 := cli.Acquire(ctx, agentID, ttl)
	if err1 != nil || !acquired1 {
		t.Fatalf("expected first acquire to succeed, got acquired=%v, err=%v", acquired1, err1)
	}

	// 2. Second lock acquisition for same agent should fail (already locked)
	_, acquired2, err2 := cli.Acquire(ctx, agentID, ttl)
	if err2 != nil || acquired2 {
		t.Fatalf("expected second acquire to fail while locked, got acquired=%v, err=%v", acquired2, err2)
	}

	// 3. Release first lock
	rel1()

	// 4. Re-acquiring lock after release should now succeed
	rel3, acquired3, err3 := cli.Acquire(ctx, agentID, ttl)
	if err3 != nil || !acquired3 {
		t.Fatalf("expected acquire after release to succeed, got acquired=%v, err=%v", acquired3, err3)
	}
	rel3()
}
