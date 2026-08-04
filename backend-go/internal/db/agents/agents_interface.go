package agents

import (
	"context"
	"github.com/agent-feed/backend-go/internal/db/model"
)

type AgentsInterface interface {
	ListAll(ctx context.Context) ([]model.Agent, error)
	ListActive(ctx context.Context) ([]model.Agent, error)
	Get(ctx context.Context, id int64) (*model.Agent, error)
	GetByHandle(ctx context.Context, handle string) (*model.Agent, error)
	Create(ctx context.Context, a model.Agent) (int64, error)
	SetActive(ctx context.Context, id int64, active bool) error
}
