// Package notifier handles notifying thread subscribers and followers when new posts or replies occur.
package notifier

import (
	"context"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/subscriptions"
	"github.com/twitter-for-ai-agents/backend-go/internal/scheduler/evaluator"
)

// Interface contract for the subscriber notifier
type NotifierInterface interface {
	NotifySubscribers(ctx context.Context, created model.Post, agentMap map[int64]model.Agent)
}

// Interface dependency on turn evaluator
type EvaluatorInterface interface {
	EvaluateAgentTurn(ctx context.Context, t evaluator.Turn)
	IsAgentActive(ctx context.Context, id int64) bool
}

// Subscriptions notifier managing subscriber alerts and recheck turn submissions
type Notifier struct {
	subs      subscriptions.SubscriptionInterface
	evaluator EvaluatorInterface
}

// Instantiates a new Notifier with subscription store and evaluator dependencies
func New(subs subscriptions.SubscriptionInterface, eval EvaluatorInterface) NotifierInterface {
	return &Notifier{
		subs:      subs,
		evaluator: eval,
	}
}

// Alerts subscribers and followers when a new post or reply is created, enqueuing recheck turns
func (n *Notifier) NotifySubscribers(ctx context.Context, created model.Post, agentMap map[int64]model.Agent) {
	if n == nil {
		return
	}
	var subscriberIDs, followerIDs []int64
	if created.ParentID != nil {
		if created.RootID != nil && n.subs != nil {
			if subs, err := n.subs.ListSubscribers(ctx, *created.RootID); err == nil {
				subscriberIDs = subs
			}
		}
	} else {
		for id := range agentMap {
			if id != created.AgentID {
				followerIDs = append(followerIDs, id)
			}
		}
	}

	rootID := created.RootID
	if rootID == nil && created.ParentID == nil {
		rootID = &created.ID
	}
	if n.subs != nil {
		for _, mentionID := range created.Mentions {
			if mentionID != created.AgentID && rootID != nil {
				_ = n.subs.Subscribe(ctx, mentionID, *rootID)
			}
		}
	}

	if n.evaluator != nil {
		for _, targetID := range selectNotifyTargets(created, subscriberIDs, followerIDs) {
			if targetAgent, ok := agentMap[targetID]; ok {
				if n.evaluator.IsAgentActive(ctx, targetID) {
					p := created
					t := evaluator.Turn{
						Mode:       evaluator.ModeRecheck,
						Agent:      targetAgent,
						AgentMap:   agentMap,
						TargetPost: &p,
					}
					n.evaluator.EvaluateAgentTurn(ctx, t)
				}
			}
		}
	}
}

// Selects target agent IDs to be notified (combining subscribers, followers, and mentions, excluding author)
func selectNotifyTargets(created model.Post, subscriberIDs, followerIDs []int64) []int64 {
	var base []int64
	if created.ParentID != nil {
		base = subscriberIDs
	} else {
		base = followerIDs
	}
	targets := make([]int64, 0, len(base)+len(created.Mentions)+1)
	targets = append(targets, base...)
	targets = append(targets, created.Mentions...)
	targets = dedupeInts(targets)

	out := targets[:0]
	for _, t := range targets {
		if t != created.AgentID {
			out = append(out, t)
		}
	}
	return out
}

// Deduplicates an array of int64 IDs
func dedupeInts(in []int64) []int64 {
	seen := make(map[int64]bool, len(in))
	out := make([]int64, 0, len(in))
	for _, v := range in {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}
