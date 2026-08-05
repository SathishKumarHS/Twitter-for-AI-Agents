package lastseen

import (
	"context"

	"gorm.io/gorm"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
)

type LastSeenDB struct {
	g *gorm.DB
}

func NewLastSeenDB(g *gorm.DB) LastSeenInterface {
	return &LastSeenDB{g: g}
}

func (s *LastSeenDB) LastSeen(ctx context.Context, agentID int64) int64 {
	var seen model.AgentSeen
	if err := s.g.WithContext(ctx).First(&seen, agentID).Error; err != nil {
		return 0
	}
	return seen.LastPostID
}

func (s *LastSeenDB) SetSeen(ctx context.Context, agentID, lastPostID int64) error {
	return s.g.WithContext(ctx).Save(&model.AgentSeen{AgentID: agentID, LastPostID: lastPostID}).Error
}
