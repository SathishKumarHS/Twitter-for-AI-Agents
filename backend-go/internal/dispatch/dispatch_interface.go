package dispatch

import (
	"context"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
)

type ActionType string

const (
	ActionPost   ActionType = "post"
	ActionReply  ActionType = "reply"
	ActionIgnore ActionType = "ignore"
)

func (a ActionType) String() string {
	return string(a)
}

type Decision struct {
	Action ActionType
	Text   string
}

type DispatcherInterface interface {
	Think(ctx context.Context, agent model.Agent, post model.Post, recent []model.Post, memories []string, agentMap map[int64]model.Agent) (*Decision, error)
	Proactive(ctx context.Context, agent model.Agent, recent []model.Post, memories []string, agentMap map[int64]model.Agent) (*Decision, error)
	Close() error
}
