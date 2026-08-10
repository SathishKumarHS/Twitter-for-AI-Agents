package redis

import (
	"context"
	"fmt"
)

const (
	feedRecentKeyFmt  = "feed:recent:%d"
	agentActiveKeyFmt = "agent:active:%d"
)

// Closes the underlying Redis connection pool
func (c *Client) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

// Retrieves cached timeline feed payload for a given limit, or nil if cache miss
func (c *Client) GetFeed(ctx context.Context, limit int) *string {
	if c == nil || c.client == nil {
		return nil
	}
	val, err := c.client.Get(ctx, fmt.Sprintf(feedRecentKeyFmt, limit)).Result()
	if err != nil || val == "" {
		return nil
	}
	return &val
}

// Stores timeline feed JSON payload in Redis with configured feedTTL
func (c *Client) SetFeed(ctx context.Context, limit int, payload []byte) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Set(ctx, fmt.Sprintf(feedRecentKeyFmt, limit), payload, c.feedTTL).Err()
}

// Removes timeline feed cache keys from Redis when new posts are created
func (c *Client) InvalidateFeed(ctx context.Context) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Del(
		ctx,
		fmt.Sprintf(feedRecentKeyFmt, 50),
		fmt.Sprintf(feedRecentKeyFmt, 100),
		fmt.Sprintf(feedRecentKeyFmt, 20),
	).Err()
}

// Retrieves cached active status for an agent ID (returns active status and found flag)
func (c *Client) IsAgentActive(ctx context.Context, id int64) (bool, bool) {
	if c == nil || c.client == nil {
		return false, false
	}
	val, err := c.client.Get(ctx, fmt.Sprintf(agentActiveKeyFmt, id)).Result()
	if err != nil {
		return false, false
	}
	return val == "1" || val == "true", true
}

// Stores an agent's active status in Redis with configured agentStatusTTL
func (c *Client) SetAgentActive(ctx context.Context, id int64, active bool) error {
	if c == nil || c.client == nil {
		return nil
	}
	val := "0"
	if active {
		val = "1"
	}
	return c.client.Set(ctx, fmt.Sprintf(agentActiveKeyFmt, id), val, c.agentStatusTTL).Err()
}
