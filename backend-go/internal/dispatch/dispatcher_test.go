package dispatch_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"

	twitterforaiworkerpb "github.com/twitter-for-ai-agents/backend-go/gen/twitterforaiworkerpb"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/internal/dispatch"
	"github.com/twitter-for-ai-agents/backend-go/mocks"
)

type actResult struct {
	decision *dispatch.Decision
	err      error
}

func awaitAct(t *testing.T, ch <-chan actResult) (*dispatch.Decision, error) {
	t.Helper()
	select {
	case res := <-ch:
		return res.decision, res.err
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for the Act reply handler")
		return nil, nil
	}
}

func TestDispatcher_ActMapsDecision(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockWorkerClient(ctrl)
	dispatcher := dispatch.NewWithClient(mockClient)

	agent := model.Agent{ID: 1, Handle: "test_bot", Name: "Test Bot"}
	// The scheduler decorates the target post's author; the dispatcher must pass
	// it through to the worker.
	post := model.Post{ID: 10, AgentID: 2, Body: "Hello", AuthorHandle: "user2", AuthorName: "User Two"}

	mockClient.EXPECT().Think(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, req *twitterforaiworkerpb.ThinkRequest, opts ...grpc.CallOption) (*twitterforaiworkerpb.ThinkResponse, error) {
			if req.Mode != "respond" {
				t.Errorf("expected mode 'respond', got '%s'", req.Mode)
			}
			if req.Agent.Handle != "test_bot" {
				t.Errorf("expected agent handle 'test_bot', got '%s'", req.Agent.Handle)
			}
			if req.Post.Id != 10 {
				t.Errorf("expected post id 10, got %d", req.Post.Id)
			}
			if req.Post.AuthorHandle != "user2" {
				t.Errorf("expected decorated author handle on the request post, got %q", req.Post.AuthorHandle)
			}
			return &twitterforaiworkerpb.ThinkResponse{
				Executed: false,
			}, nil
		},
	).Times(1)

	ch := make(chan actResult, 1)
	dispatcher.Act(context.Background(), dispatch.ModeRespond, agent, post, func(decision *dispatch.Decision, err error) {
		ch <- actResult{decision: decision, err: err}
	})

	decision, err := awaitAct(t, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Executed {
		t.Errorf("expected Executed=false, got true")
	}
}

func TestDispatcher_ActSurfacesExecuted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockWorkerClient(ctrl)
	dispatcher := dispatch.NewWithClient(mockClient)

	agent := model.Agent{ID: 1, Handle: "test_bot", Name: "Test Bot"}

	mockClient.EXPECT().Think(gomock.Any(), gomock.Any()).Return(&twitterforaiworkerpb.ThinkResponse{
		Executed: true,
	}, nil)

	ch := make(chan actResult, 1)
	dispatcher.Act(context.Background(), dispatch.ModeProactive, agent, model.Post{}, func(decision *dispatch.Decision, err error) {
		ch <- actResult{decision: decision, err: err}
	})

	decision, err := awaitAct(t, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Executed {
		t.Fatalf("expected Executed=true when the worker already acted")
	}
}

func TestDispatcher_ActError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockWorkerClient(ctrl)
	dispatcher := dispatch.NewWithClient(mockClient)

	agent := model.Agent{ID: 1, Handle: "test_bot", Name: "Test Bot"}

	mockClient.EXPECT().Think(gomock.Any(), gomock.Any()).Return(nil, errors.New("gRPC connection failed"))

	ch := make(chan actResult, 1)
	dispatcher.Act(context.Background(), dispatch.ModeRespond, agent, model.Post{}, func(decision *dispatch.Decision, err error) {
		ch <- actResult{decision: decision, err: err}
	})

	decision, err := awaitAct(t, ch)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if decision != nil {
		t.Errorf("expected nil decision on error, got %v", decision)
	}
}
