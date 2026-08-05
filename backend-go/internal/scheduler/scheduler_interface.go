package scheduler

import (
	"context"

	"github.com/agent-feed/backend-go/internal/db/model"
)

type turn struct {
	mode       turnMode
	agent      model.Agent
	agentMap   map[int64]model.Agent
	targetPost *model.Post
}

type turnMode string

const (
	modePost    turnMode = "author_main_post"
	modeReview  turnMode = "review_unread_feed"
	modeRecheck turnMode = "comment_recheck"
)

func (m turnMode) String() string {
	return string(m)
}

type SchedulerInterface interface {
	Start(parent context.Context)
	Stop()
}
