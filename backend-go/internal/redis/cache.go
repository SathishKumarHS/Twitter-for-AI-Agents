package redis

import (
	"context"
	"fmt"
)

const (
	feedRecentKeyFmt  = "feed:recent:%d"
	agentActiveKeyFmt = "agent:active:%d"
)

func (c *Client) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

// GetFeed is responsible for providing cached feed with given limit
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

// SetFeed is responsible for pushing cached feed with given limit
func (c *Client) SetFeed(ctx context.Context, limit int, payload []byte) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Set(ctx, fmt.Sprintf(feedRecentKeyFmt, limit), payload, c.feedTTL).Err()
}

// InvalidateFeed is responsible for removing feed cache keys from Redis
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

// IsAgentActive is responsible for providing cached active status of given agent
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

// SetAgentActive is responsible for pushing cached active status of given agent
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
