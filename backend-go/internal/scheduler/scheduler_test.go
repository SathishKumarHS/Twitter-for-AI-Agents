package scheduler

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/twitter-for-ai-agents/backend-go/internal/db"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/internal/dispatch"
	"github.com/twitter-for-ai-agents/backend-go/mocks"
)

func TestSchedulerStartStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgents := mocks.NewMockAgentsInterface(ctrl)
	mockPosts := mocks.NewMockPostInterface(ctrl)
	mockLastSeen := mocks.NewMockLastSeenInterface(ctrl)
	mockMemories := mocks.NewMockMemoriesInterface(ctrl)
	mockDispatcher := mocks.NewMockDispatcherInterface(ctrl)
	mockHub := mocks.NewMockHubInterface(ctrl)
	mockRedis := mocks.NewMockRedisInterface(ctrl)

	mockAgents.EXPECT().ListActive(gomock.Any()).Return([]model.Agent{
		{ID: 1, Handle: "agent1", Name: "Agent One", Active: true},
	}, nil).AnyTimes()

	mockPosts.EXPECT().LastMainPost(gomock.Any(), gomock.Any()).Return(time.Time{}, nil).AnyTimes()
	mockPosts.EXPECT().NewPostsAfter(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockPosts.EXPECT().Recent(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	mockLastSeen.EXPECT().LastSeen(gomock.Any(), gomock.Any()).Return(int64(0)).AnyTimes()
	mockMemories.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	mockDispatcher.EXPECT().Proactive(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&dispatch.Decision{Action: dispatch.ActionIgnore}, nil).AnyTimes()
	mockDispatcher.EXPECT().Think(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&dispatch.Decision{Action: dispatch.ActionIgnore}, nil).AnyTimes()

	mockRedis.EXPECT().Acquire(gomock.Any(), gomock.Any(), gomock.Any()).Return(func() {}, true, nil).AnyTimes()
	mockRedis.EXPECT().IsAgentActive(gomock.Any(), gomock.Any()).Return(true, true).AnyTimes()

	database := &db.DB{
		AgentsDB:   mockAgents,
		PostDB:     mockPosts,
		LastSeenDB: mockLastSeen,
		MemoriesDB: mockMemories,
	}

	sched := New(
		100*time.Millisecond,
		1*time.Second,
		1*time.Second,
		2,
		mockRedis,
		database,
		mockDispatcher,
		mockHub,
		mockRedis,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sched.Start(ctx)

	time.Sleep(150 * time.Millisecond)

	sched.Stop()
}
