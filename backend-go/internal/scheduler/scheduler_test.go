package scheduler_test

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/twitter-for-ai-agents/backend-go/internal/db"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/internal/scheduler"
	"github.com/twitter-for-ai-agents/backend-go/internal/scheduler/evaluator"
	"github.com/twitter-for-ai-agents/backend-go/mocks"
)

type fakeEvaluator struct {
	evaluated []evaluator.Turn
}

func (f *fakeEvaluator) EvaluateAgentTurn(_ context.Context, t evaluator.Turn) {
	f.evaluated = append(f.evaluated, t)
}

func (f *fakeEvaluator) IsAgentActive(_ context.Context, _ int64) bool {
	return true
}

func TestSchedulerStartStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgents := mocks.NewMockAgentsInterface(ctrl)

	mockAgents.EXPECT().ListActive(gomock.Any()).Return([]model.Agent{
		{ID: 1, Handle: "agent1", Name: "Agent One", Active: true},
	}, nil).AnyTimes()

	database := &db.DB{
		AgentsDB: mockAgents,
	}
	eval := &fakeEvaluator{}

	sched := scheduler.New(
		50*time.Millisecond,
		100*time.Millisecond,
		2,
		database,
		eval,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sched.Start(ctx)
	time.Sleep(120 * time.Millisecond)
	sched.Stop()
}
