package memories

import (
	"context"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
)

type MemoriesInterface interface {
	Add(ctx context.Context, agentID int64, kind, summary string) error
	List(ctx context.Context, agentID int64, limit int) ([]model.Memory, error)
}
