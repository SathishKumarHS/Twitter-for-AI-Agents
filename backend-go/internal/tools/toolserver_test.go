package tools

import (
	"context"
	"encoding/json"
	"testing"

	"go.uber.org/mock/gomock"

	twitterforaiworkerpb "github.com/twitter-for-ai-agents/backend-go/gen/twitterforaiworkerpb"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/mocks"
)

func newToolServer(ctrl *gomock.Controller) (*toolServer, *mocks.MockAgentsInterface, *mocks.MockPostInterface, *mocks.MockMemoriesInterface) {
	mockAgents := mocks.NewMockAgentsInterface(ctrl)
	mockPosts := mocks.NewMockPostInterface(ctrl)
	mockMemories := mocks.NewMockMemoriesInterface(ctrl)
	srv := NewToolGRPCServer(mockAgents, mockPosts, mockMemories, nil, nil, nil, nil, nil, 10, 20)
	return srv.(*toolServer), mockAgents, mockPosts, mockMemories
}

func TestToolServerExecuteGetFeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	srv, mockAgents, mockPosts, _ := newToolServer(ctrl)
	agent := model.Agent{ID: 1, Handle: "agent1", Name: "Agent One", Active: true}

	mockAgents.EXPECT().Get(gomock.Any(), int64(1)).Return(&agent, nil).Times(1)
	mockAgents.EXPECT().ListActive(gomock.Any()).Return([]model.Agent{agent}, nil).Times(1)
	mockPosts.EXPECT().Recent(gomock.Any(), gomock.Any()).Return([]model.Post{{ID: 1, AgentID: 1, Body: "hi"}}, nil).Times(1)

	resp, err := srv.ExecuteTool(context.Background(), &twitterforaiworkerpb.ToolCall{AgentId: 1, Name: "get_feed", Arguments: `{"limit": 10}`})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Error {
		t.Fatalf("expected no error, got: %s", resp.Output)
	}
	var posts []map[string]any
	if err := json.Unmarshal([]byte(resp.Output), &posts); err != nil {
		t.Fatalf("get_feed output is not a JSON array: %v", err)
	}
	if len(posts) != 1 || posts[0]["handle"] != "agent1" {
		t.Errorf("unexpected feed output: %s", resp.Output)
	}
	if resp.Acted {
		t.Errorf("get_feed is a read tool; expected acted=false")
	}
}

func TestToolServerExecuteUnknownTool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	srv, mockAgents, _, _ := newToolServer(ctrl)
	agent := model.Agent{ID: 1, Handle: "agent1", Name: "Agent One", Active: true}

	mockAgents.EXPECT().Get(gomock.Any(), int64(1)).Return(&agent, nil).Times(1)
	mockAgents.EXPECT().ListActive(gomock.Any()).Return([]model.Agent{agent}, nil).Times(1)

	resp, err := srv.ExecuteTool(context.Background(), &twitterforaiworkerpb.ToolCall{AgentId: 1, Name: "drop_database"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Error {
		t.Fatalf("expected an error result for an unknown tool, got: %s", resp.Output)
	}
}

func TestToolServerExecuteInactiveAgent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	srv, mockAgents, _, _ := newToolServer(ctrl)
	mockAgents.EXPECT().Get(gomock.Any(), int64(1)).Return(&model.Agent{ID: 1, Handle: "agent1", Active: false}, nil).Times(1)

	resp, err := srv.ExecuteTool(context.Background(), &twitterforaiworkerpb.ToolCall{AgentId: 1, Name: "get_feed"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Error {
		t.Fatalf("expected an error result for an inactive agent, got: %s", resp.Output)
	}
}

func TestToolServerExecutePostActs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	srv, mockAgents, mockPosts, mockMemories := newToolServer(ctrl)
	agent := model.Agent{ID: 1, Handle: "agent1", Name: "Agent One", Active: true}

	mockAgents.EXPECT().Get(gomock.Any(), int64(1)).Return(&agent, nil).Times(1)
	mockAgents.EXPECT().ListActive(gomock.Any()).Return([]model.Agent{agent}, nil).Times(1)
	mockPosts.EXPECT().Create(gomock.Any(), int64(1), "Hello", int64(0)).Return(model.Post{ID: 9, AgentID: 1, Body: "Hello", Kind: "post"}, nil).Times(1)
	mockMemories.EXPECT().Add(gomock.Any(), int64(1), "learning", "posted: Hello").Return(nil).Times(1)

	resp, err := srv.ExecuteTool(context.Background(), &twitterforaiworkerpb.ToolCall{AgentId: 1, Name: "post", Arguments: `{"text": "Hello"}`})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Error {
		t.Fatalf("expected no error, got: %s", resp.Output)
	}
	if !resp.Acted {
		t.Errorf("post is a write tool; expected acted=true")
	}
}

func TestToolServerExecutePostRejectedInRespondMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	srv, mockAgents, _, _ := newToolServer(ctrl)
	agent := model.Agent{ID: 1, Handle: "agent1", Name: "Agent One", Active: true}

	mockAgents.EXPECT().Get(gomock.Any(), int64(1)).Return(&agent, nil).Times(1)
	mockAgents.EXPECT().ListActive(gomock.Any()).Return([]model.Agent{agent}, nil).Times(1)

	resp, err := srv.ExecuteTool(context.Background(), &twitterforaiworkerpb.ToolCall{
		AgentId: 1, Name: "post", Arguments: `{"text": "nope"}`, Mode: "respond",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Error {
		t.Fatalf("expected mode error for post in respond turn, got: %s", resp.Output)
	}
	if resp.Acted {
		t.Fatalf("post must not act in respond mode")
	}
}
