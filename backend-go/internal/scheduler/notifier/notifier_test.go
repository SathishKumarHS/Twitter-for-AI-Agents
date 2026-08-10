package notifier_test

import (
	"context"
	"testing"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/internal/scheduler/evaluator"
	"github.com/twitter-for-ai-agents/backend-go/internal/scheduler/notifier"
)

type fakeSubs struct {
	subs       map[int64][]int64
	subscribed []int64
}

func (f *fakeSubs) ListSubscribers(_ context.Context, postID int64) ([]int64, error) {
	return f.subs[postID], nil
}

func (f *fakeSubs) Subscribe(_ context.Context, agentID, _ int64) error {
	f.subscribed = append(f.subscribed, agentID)
	return nil
}

type fakeEvaluator struct {
	evaluated []int64
}

func (f *fakeEvaluator) EvaluateAgentTurn(_ context.Context, t evaluator.Turn) {
	f.evaluated = append(f.evaluated, t.Agent.ID)
}

func (f *fakeEvaluator) IsAgentActive(_ context.Context, _ int64) bool {
	return true
}

func TestNotifier_NotifySubscribers_MainPost(t *testing.T) {
	subs := &fakeSubs{}
	eval := &fakeEvaluator{}
	n := notifier.New(subs, eval)

	agentMap := map[int64]model.Agent{
		1: {ID: 1, Handle: "author", Active: true},
		2: {ID: 2, Handle: "follower", Active: true},
	}

	post := model.Post{ID: 10, AgentID: 1, Body: "Hello"}
	n.NotifySubscribers(context.Background(), post, agentMap)

	if len(eval.evaluated) != 1 || eval.evaluated[0] != 2 {
		t.Fatalf("expected follower (ID 2) to be evaluated, got: %v", eval.evaluated)
	}
}

func TestNotifier_NotifySubscribers_Reply(t *testing.T) {
	rootID := int64(10)
	parentID := int64(10)
	subs := &fakeSubs{
		subs: map[int64][]int64{10: {1, 3}},
	}
	eval := &fakeEvaluator{}
	n := notifier.New(subs, eval)

	agentMap := map[int64]model.Agent{
		1: {ID: 1, Handle: "author", Active: true},
		2: {ID: 2, Handle: "replier", Active: true},
		3: {ID: 3, Handle: "subscriber", Active: true},
	}

	reply := model.Post{ID: 20, AgentID: 2, ParentID: &parentID, RootID: &rootID, Body: "Reply text"}
	n.NotifySubscribers(context.Background(), reply, agentMap)

	if len(eval.evaluated) != 2 || eval.evaluated[0] != 1 || eval.evaluated[1] != 3 {
		t.Fatalf("expected subscribers (IDs 1, 3) to be evaluated, got: %v", eval.evaluated)
	}
}
