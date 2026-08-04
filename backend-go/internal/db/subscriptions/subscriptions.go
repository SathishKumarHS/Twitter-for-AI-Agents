package subscriptions

import (
	"context"

	"github.com/agent-feed/backend-go/internal/db/model"
	"gorm.io/gorm"
)

type SubcriptionDB struct {
	g *gorm.DB
}

func NewSubcriptionsDB(g *gorm.DB) SubscriptionInterface {
	return &SubcriptionDB{g: g}
}

func (s *SubcriptionDB) Subscribe(ctx context.Context, agentID, rootPostID int64) error {
	return s.g.WithContext(ctx).
		Where("agent_id = ? AND root_post_id = ?", agentID, rootPostID).
		FirstOrCreate(&model.ThreadSubscriber{AgentID: agentID, RootPostID: rootPostID}).
		Error
}

func (s *SubcriptionDB) ListSubscribers(ctx context.Context, rootPostID int64) ([]int64, error) {
	var ids []int64
	err := s.g.WithContext(ctx).
		Model(&model.ThreadSubscriber{}).
		Where("root_post_id = ?", rootPostID).
		Order("agent_id").
		Pluck("agent_id", &ids).Error
	return ids, err
}
