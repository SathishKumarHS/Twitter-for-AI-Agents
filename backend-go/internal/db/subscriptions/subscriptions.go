// Package subscriptions implements storage and query operations for thread subscribers.
package subscriptions

import (
	"context"

	"gorm.io/gorm"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
)

// GORM-backed storage implementation for thread subscriptions
type SubscriptionDB struct {
	g *gorm.DB
}

// Instantiates a new SubscriptionDB store implementing SubscriptionInterface
func NewSubscriptionDB(g *gorm.DB) SubscriptionInterface {
	return &SubscriptionDB{g: g}
}

// Subscribes an agent to a conversation thread if not already subscribed
func (s *SubscriptionDB) Subscribe(ctx context.Context, agentID, rootPostID int64) error {
	return s.g.WithContext(ctx).
		Where("agent_id = ? AND root_post_id = ?", agentID, rootPostID).
		FirstOrCreate(&model.ThreadSubscriber{AgentID: agentID, RootPostID: rootPostID}).
		Error
}

// Lists all subscriber agent IDs subscribed to a specified root post thread
func (s *SubscriptionDB) ListSubscribers(ctx context.Context, rootPostID int64) ([]int64, error) {
	var ids []int64
	err := s.g.WithContext(ctx).
		Model(&model.ThreadSubscriber{}).
		Where("root_post_id = ?", rootPostID).
		Order("agent_id").
		Pluck("agent_id", &ids).Error
	return ids, err
}
