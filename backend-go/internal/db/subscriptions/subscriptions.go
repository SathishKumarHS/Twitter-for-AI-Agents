// Package subscriptions tracks which agents are subscribed to a thread and
// should be notified when a new comment is posted inside it.
package subscriptions

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// ThreadSubscriber is an agent subscribed to a thread (root post).
type ThreadSubscriber struct {
	AgentID    int64     `gorm:"primaryKey" json:"agent_id"`
	RootPostID int64     `gorm:"primaryKey" json:"root_post_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// TableName gives GORM its Postgres table name.
func (ThreadSubscriber) TableName() string { return "thread_subscribers" }

// Store is the storage contract for thread subscriptions.
type Store interface {
	Subscribe(ctx context.Context, agentID, rootPostID int64) error
	Subscribers(ctx context.Context, rootPostID int64) ([]int64, error)
}

// GormStore is the GORM-backed implementation of Store.
type GormStore struct {
	g *gorm.DB
}

// NewGorm returns a Store backed by GORM.
func NewGorm(g *gorm.DB) Store {
	return &GormStore{g: g}
}

func (s *GormStore) Subscribe(ctx context.Context, agentID, rootPostID int64) error {
	return s.g.WithContext(ctx).
		Where("agent_id = ? AND root_post_id = ?", agentID, rootPostID).
		FirstOrCreate(&ThreadSubscriber{AgentID: agentID, RootPostID: rootPostID}).
		Error
}

func (s *GormStore) Subscribers(ctx context.Context, rootPostID int64) ([]int64, error) {
	var ids []int64
	err := s.g.WithContext(ctx).
		Model(&ThreadSubscriber{}).
		Where("root_post_id = ?", rootPostID).
		Order("agent_id").
		Pluck("agent_id", &ids).Error
	return ids, err
}
