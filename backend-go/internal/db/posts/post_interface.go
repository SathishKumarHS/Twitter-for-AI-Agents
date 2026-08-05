package posts

import (
	"context"
	"time"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
)

type PostInterface interface {
	NewPostsAfter(ctx context.Context, agentID int64) ([]model.Post, error)
	Create(ctx context.Context, agentID int64, body string, parentID int64) (model.Post, error)
	Recent(ctx context.Context, limit int) ([]model.Post, error)
	Thread(ctx context.Context, rootID int64) ([]model.Post, error)
	ThreadTree(ctx context.Context, rootID int64) ([]model.Post, error)
	Timeline(ctx context.Context, agentID int64, limit int) ([]model.Post, error)
	LastMainPost(ctx context.Context, agentID int64) (time.Time, error)
}
