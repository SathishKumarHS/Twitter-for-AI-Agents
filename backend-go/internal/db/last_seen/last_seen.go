// Package lastseen implements tracking for each agent's last-seen post ID.
package lastseen

import (
	"context"

	"gorm.io/gorm"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
)

// GORM-backed storage implementation for agent last-seen post IDs
type LastSeenDB struct {
	g *gorm.DB
}

// Instantiates a new LastSeenDB store implementing LastSeenInterface
func NewLastSeenDB(g *gorm.DB) LastSeenInterface {
	return &LastSeenDB{g: g}
}

// Retrieves the highest post ID seen by the specified agent, or 0 if none recorded
func (s *LastSeenDB) LastSeen(ctx context.Context, agentID int64) int64 {
	var seen model.AgentSeen
	if err := s.g.WithContext(ctx).First(&seen, agentID).Error; err != nil {
		return 0
	}
	return seen.LastPostID
}

// Updates or creates the last-seen post ID record for a specified agent
func (s *LastSeenDB) SetSeen(ctx context.Context, agentID, lastPostID int64) error {
	return s.g.WithContext(ctx).Save(&model.AgentSeen{AgentID: agentID, LastPostID: lastPostID}).Error
}
