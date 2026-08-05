package dispatch_test

import (
	"context"
	"errors"
	"testing"

	twitterforaiworkerpb "github.com/twitter-for-ai-agents/backend-go/gen/twitterforaiworkerpb"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/internal/dispatch"
	"github.com/twitter-for-ai-agents/backend-go/mocks"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
)

func TestDispatcher_ThinkSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockWorkerClient(ctrl)
	dispatcher := dispatch.NewWithClient(mockClient)

	agent := model.Agent{ID: 1, Handle: "test_bot", Name: "Test Bot"}
	post := model.Post{ID: 10, AgentID: 2, Body: "Hello world"}
	recent := []model.Post{{ID: 5, AgentID: 1, Body: "Previous post"}}
	memories := []string{"memory 1"}
	agentMap := map[int64]model.Agent{
		1: agent,
		2: {ID: 2, Handle: "user2", Name: "User Two"},
	}

	mockClient.EXPECT().Think(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, req *twitterforaiworkerpb.ThinkRequest, opts ...grpc.CallOption) (*twitterforaiworkerpb.ThinkReply, error) {
			if req.Mode != "respond" {
				t.Errorf("expected mode 'respond', got '%s'", req.Mode)
			}
			if req.Agent.Handle != "test_bot" {
				t.Errorf("expected agent handle 'test_bot', got '%s'", req.Agent.Handle)
			}
			if req.Post.Id != 10 {
				t.Errorf("expected post id 10, got %d", req.Post.Id)
			}
			if len(req.Recent) != 1 || req.Recent[0].Id != 5 {
				t.Errorf("unexpected recent posts in request")
			}
			if len(req.Memories) != 1 || req.Memories[0] != "memory 1" {
				t.Errorf("unexpected memories in request")
			}

			return &twitterforaiworkerpb.ThinkReply{
				Action: "reply",
				Text:   "Nice to meet you!",
			}, nil
		},
	)

	decision, err := dispatcher.Think(context.Background(), agent, post, recent, memories, agentMap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decision.Action != dispatch.ActionReply {
		t.Errorf("expected action 'reply', got '%s'", decision.Action)
	}
	if decision.Text != "Nice to meet you!" {
		t.Errorf("expected text 'Nice to meet you!', got '%s'", decision.Text)
	}
}

func TestDispatcher_ProactiveSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockWorkerClient(ctrl)
	dispatcher := dispatch.NewWithClient(mockClient)

	agent := model.Agent{ID: 1, Handle: "test_bot", Name: "Test Bot"}
	recent := []model.Post{}
	memories := []string{}
	agentMap := map[int64]model.Agent{1: agent}

	mockClient.EXPECT().Think(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, req *twitterforaiworkerpb.ThinkRequest, opts ...grpc.CallOption) (*twitterforaiworkerpb.ThinkReply, error) {
			if req.Mode != "proactive" {
				t.Errorf("expected mode 'proactive', got '%s'", req.Mode)
			}
			return &twitterforaiworkerpb.ThinkReply{
				Action: "post",
				Text:   "Proactive AI tweet!",
			}, nil
		},
	)

	decision, err := dispatcher.Proactive(context.Background(), agent, recent, memories, agentMap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decision.Action != dispatch.ActionPost {
		t.Errorf("expected action 'post', got '%s'", decision.Action)
	}
	if decision.Text != "Proactive AI tweet!" {
		t.Errorf("expected text 'Proactive AI tweet!', got '%s'", decision.Text)
	}
}

func TestDispatcher_ThinkError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockWorkerClient(ctrl)
	dispatcher := dispatch.NewWithClient(mockClient)

	agent := model.Agent{ID: 1, Handle: "test_bot", Name: "Test Bot"}

	mockClient.EXPECT().Think(gomock.Any(), gomock.Any()).Return(nil, errors.New("gRPC connection failed"))

	decision, err := dispatcher.Think(context.Background(), agent, model.Post{}, nil, nil, nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if decision != nil {
		t.Errorf("expected nil decision on error, got %v", decision)
	}
}
