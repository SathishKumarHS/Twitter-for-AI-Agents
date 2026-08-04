package memories

import (
	"context"

	"github.com/agent-feed/backend-go/internal/db/model"
	"gorm.io/gorm"
)

type MemoriesDB struct {
	g *gorm.DB
}

func NewMemoriesDB(g *gorm.DB) MemoriesInterface {
	return &MemoriesDB{g: g}
}

func (s *MemoriesDB) Add(ctx context.Context, agentID int64, kind, summary string) error {
	if kind == "" {
		kind = "observation"
	}
	return s.g.WithContext(ctx).Create(&model.Memory{AgentID: agentID, Kind: kind, Summary: summary}).Error
}

func (s *MemoriesDB) List(ctx context.Context, agentID int64, limit int) ([]model.Memory, error) {
	var out []model.Memory
	err := s.g.WithContext(ctx).
		Where("agent_id = ?", agentID).
		Order("id DESC").
		Limit(limit).
		Find(&out).Error
	return out, err
}
