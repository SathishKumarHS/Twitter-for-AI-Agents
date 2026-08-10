package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/mocks"
)

func setupExecutor(ctrl *gomock.Controller, mode string) (
	ExecutorInterface,
	*mocks.MockPostInterface,
	*mocks.MockMemoriesInterface,
	*mocks.MockLastSeenInterface,
	*mocks.MockSubscriptionInterface,
	*mocks.MockHubInterface,
	*mocks.MockRedisInterface,
) {
	mockPosts := mocks.NewMockPostInterface(ctrl)
	mockMemories := mocks.NewMockMemoriesInterface(ctrl)
	mockLastSeen := mocks.NewMockLastSeenInterface(ctrl)
	mockSubs := mocks.NewMockSubscriptionInterface(ctrl)
	mockHub := mocks.NewMockHubInterface(ctrl)
	mockRedis := mocks.NewMockRedisInterface(ctrl)

	agent := model.Agent{ID: 1, Handle: "agent1", Name: "Agent One"}
	agentMap := map[int64]model.Agent{
		1: agent,
		2: {ID: 2, Handle: "agent2", Name: "Agent Two"},
	}

	exec := New(
		mockPosts,
		mockMemories,
		mockLastSeen,
		mockSubs,
		mockHub,
		mockRedis,
		nil, // notifier
		agent,
		agentMap,
		10, // default feed limit
		20, // max feed limit
		mode,
	)
	return exec, mockPosts, mockMemories, mockLastSeen, mockSubs, mockHub, mockRedis
}

func TestExecutorExecuteUnknownTool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	exec, _, _, _, _, _, _ := setupExecutor(ctrl, "proactive")
	res, err := exec.Execute(context.Background(), "invalid_tool", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Error || !strings.Contains(res.Output, "unknown tool") {
		t.Errorf("expected unknown tool error, got: %s", res.Output)
	}
}

func TestExecutorGetFeedLimitCaps(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	exec, mockPosts, _, _, _, _, _ := setupExecutor(ctrl, "proactive")

	// Limit requested = 100, should be capped to feedMaxSize = 20
	mockPosts.EXPECT().Recent(gomock.Any(), 20).Return([]model.Post{
		{ID: 1, AgentID: 1, Body: "Post 1"},
	}, nil).Times(1)

	res, err := exec.Execute(context.Background(), "get_feed", []byte(`{"limit": 100}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Error {
		t.Fatalf("unexpected tool error: %s", res.Output)
	}

	var views []postView
	if err := json.Unmarshal([]byte(res.Output), &views); err != nil {
		t.Fatalf("failed to parse output json: %v", err)
	}
	if len(views) != 1 || views[0].Handle != "agent1" {
		t.Errorf("unexpected output: %s", res.Output)
	}
}

func TestExecutorGetThreadValidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	exec, mockPosts, _, _, _, _, _ := setupExecutor(ctrl, "proactive")

	// Invalid root_id <= 0
	res, err := exec.Execute(context.Background(), "get_thread", []byte(`{"root_id": 0}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Error || !strings.Contains(res.Output, "root_id is required") {
		t.Errorf("expected root_id error, got: %s", res.Output)
	}

	// Valid root_id
	rootID := int64(10)
	mockPosts.EXPECT().Thread(gomock.Any(), rootID).Return([]model.Post{
		{ID: 10, AgentID: 1, Body: "Root Post"},
		{ID: 11, AgentID: 2, Body: "Reply Post", ParentID: &rootID, RootID: &rootID},
	}, nil).Times(1)

	res, err = exec.Execute(context.Background(), "get_thread", []byte(`{"root_id": 10}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Error {
		t.Fatalf("unexpected tool error: %s", res.Output)
	}

	var views []postView
	if err := json.Unmarshal([]byte(res.Output), &views); err != nil {
		t.Fatalf("failed to parse thread json: %v", err)
	}
	if len(views) != 2 || views[1].Handle != "agent2" {
		t.Errorf("unexpected thread output: %s", res.Output)
	}
}

func TestExecutorGetMemories(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	exec, _, mockMemories, _, _, _, _ := setupExecutor(ctrl, "proactive")

	mockMemories.EXPECT().List(gomock.Any(), int64(1), 10).Return([]model.Memory{
		{ID: 1, AgentID: 1, Summary: "Learned go"},
		{ID: 2, AgentID: 1, Summary: "Posted tweet"},
	}, nil).Times(1)

	res, err := exec.Execute(context.Background(), "get_memories", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Error {
		t.Fatalf("unexpected tool error: %s", res.Output)
	}

	var mems []string
	if err := json.Unmarshal([]byte(res.Output), &mems); err != nil {
		t.Fatalf("failed to parse memories json: %v", err)
	}
	if len(mems) != 2 || mems[0] != "Learned go" {
		t.Errorf("unexpected memories output: %s", res.Output)
	}
}

func TestExecutorPostModeRestriction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// When mode is "respond", post is disallowed
	exec, _, _, _, _, _, _ := setupExecutor(ctrl, "respond")

	res, err := exec.Execute(context.Background(), "post", []byte(`{"text": "Hello"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Error || !strings.Contains(res.Output, "post is not allowed while replying") {
		t.Errorf("expected post mode restriction error, got: %s", res.Output)
	}
}

func TestExecutorReplyModeRestriction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// When mode is "proactive", reply is disallowed
	exec, _, _, _, _, _, _ := setupExecutor(ctrl, "proactive")

	res, err := exec.Execute(context.Background(), "reply", []byte(`{"parent_id": 1, "text": "Hello"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Error || !strings.Contains(res.Output, "reply is not allowed while authoring") {
		t.Errorf("expected reply mode restriction error, got: %s", res.Output)
	}
}

func TestExecutorPostSuccessAndCleaning(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	exec, mockPosts, mockMemories, mockLastSeen, mockSubs, mockHub, mockRedis := setupExecutor(ctrl, "proactive")

	// Text with leading '...' marker to test cleanText
	cleanedText := "Hello world post!"

	mockPosts.EXPECT().Create(gomock.Any(), int64(1), cleanedText, int64(0)).Return(model.Post{
		ID: 100, AgentID: 1, Body: cleanedText, Kind: "post",
	}, nil).Times(1)
	mockLastSeen.EXPECT().SetSeen(gomock.Any(), int64(1), int64(100)).Return(nil).Times(1)
	mockSubs.EXPECT().Subscribe(gomock.Any(), int64(1), int64(100)).Return(nil).Times(1)
	mockHub.EXPECT().Publish(gomock.Any()).Times(1)
	mockRedis.EXPECT().InvalidateFeed(gomock.Any()).Return(nil).Times(1)
	mockMemories.EXPECT().Add(gomock.Any(), int64(1), "learning", "posted: "+cleanedText).Return(nil).Times(1)

	res, err := exec.Execute(context.Background(), "post", []byte(`{"text": "... Hello world post!"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Error || !res.Acted {
		t.Errorf("expected acted=true and no error, got: error=%v, output=%s", res.Error, res.Output)
	}
}

func TestExecutorReplySuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	exec, mockPosts, mockMemories, _, mockSubs, mockHub, mockRedis := setupExecutor(ctrl, "respond")

	parentID := int64(10)
	rootID := int64(10)
	replyText := "Nice post!"

	mockPosts.EXPECT().Create(gomock.Any(), int64(1), replyText, parentID).Return(model.Post{
		ID: 101, AgentID: 1, Body: replyText, Kind: "reply", ParentID: &parentID, RootID: &rootID,
	}, nil).Times(1)
	mockSubs.EXPECT().Subscribe(gomock.Any(), int64(1), rootID).Return(nil).Times(1)
	mockHub.EXPECT().Publish(gomock.Any()).Times(1)
	mockRedis.EXPECT().InvalidateFeed(gomock.Any()).Return(nil).Times(1)
	mockMemories.EXPECT().Add(gomock.Any(), int64(1), "learning", "replied to post #10").Return(nil).Times(1)

	res, err := exec.Execute(context.Background(), "reply", []byte(`{"parent_id": 10, "text": "Nice post!"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Error || !res.Acted {
		t.Errorf("expected acted=true and no error, got: error=%v, output=%s", res.Error, res.Output)
	}
}
