package evaluator

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/twitter-for-ai-agents/backend-go/internal/db"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/internal/redis"
	"github.com/twitter-for-ai-agents/backend-go/mocks"
)

func newTestEvaluator(cache redis.RedisInterface) *Evaluator {
	return &Evaluator{
		cache: cache,
	}
}

func TestIsAgentActiveCacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRedis := mocks.NewMockRedisInterface(ctrl)

	e := newTestEvaluator(mockRedis)
	ctx := context.Background()

	mockRedis.EXPECT().IsAgentActive(gomock.Any(), int64(1)).Return(false, true)
	if active := e.IsAgentActive(ctx, 1); active {
		t.Errorf("expected inactive from cache, got active")
	}
}

func TestIsAgentActiveDBFallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRedis := mocks.NewMockRedisInterface(ctrl)
	mockAgents := mocks.NewMockAgentsInterface(ctrl)

	e := &Evaluator{
		db:    &db.DB{AgentsDB: mockAgents},
		cache: mockRedis,
	}
	ctx := context.Background()

	mockRedis.EXPECT().IsAgentActive(gomock.Any(), int64(1)).Return(false, false)
	mockAgents.EXPECT().Get(gomock.Any(), int64(1)).Return(&model.Agent{ID: 1, Active: true}, nil)
	mockRedis.EXPECT().SetAgentActive(gomock.Any(), int64(1), true).Return(nil)

	if active := e.IsAgentActive(ctx, 1); !active {
		t.Errorf("expected active from DB fallback, got inactive")
	}
}
