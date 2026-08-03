// Package memories defines the Memory model and its storage contract.
package memories

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Memory is one long-term episodic record for an agent.
type Memory struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	AgentID   int64     `gorm:"index;not null" json:"agent_id"`
	Kind      string    `gorm:"size:32;default:'observation'" json:"kind"`
	Summary   string    `gorm:"type:text" json:"summary"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName gives GORM its Postgres table name.
func (Memory) TableName() string { return "memories" }

// Store is the storage contract for agent memories.
type Store interface {
	Add(ctx context.Context, agentID int64, kind, summary string, importance int) error
	List(ctx context.Context, agentID int64, limit int) ([]Memory, error)
}

// GormStore is the GORM-backed implementation of Store.
type GormStore struct {
	g *gorm.DB
}

// NewGorm returns a Store backed by GORM.
func NewGorm(g *gorm.DB) Store {
	return &GormStore{g: g}
}

func (s *GormStore) Add(ctx context.Context, agentID int64, kind, summary string, importance int) error {
	if kind == "" {
		kind = "observation"
	}
	return s.g.WithContext(ctx).Create(&Memory{AgentID: agentID, Kind: kind, Summary: summary}).Error
}

func (s *GormStore) List(ctx context.Context, agentID int64, limit int) ([]Memory, error) {
	var out []Memory
	err := s.g.WithContext(ctx).
		Where("agent_id = ?", agentID).
		Order("id DESC").
		Limit(limit).
		Find(&out).Error
	return out, err
}
