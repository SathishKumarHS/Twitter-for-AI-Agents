package scheduler

import (
	"context"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
)

func (s *Scheduler) notifySubscribers(ctx context.Context, created model.Post, agentMap map[int64]model.Agent) {
	var subscriberIDs, followerIDs []int64
	if created.ParentID != nil {
		if created.RootID != nil {
			if subs, err := s.db.SubscriptionsDB.ListSubscribers(ctx, *created.RootID); err == nil {
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
	for _, mentionID := range created.Mentions {
		if mentionID != created.AgentID && rootID != nil {
			_ = s.db.SubscriptionsDB.Subscribe(ctx, mentionID, *rootID)
		}
	}

	for _, targetID := range selectNotifyTargets(created, subscriberIDs, followerIDs) {
		if targetAgent, ok := agentMap[targetID]; ok {
			if s.isAgentActive(ctx, targetID) {
				p := created
				s.submitTurn(ctx, modeRecheck, targetAgent, agentMap, &p)
			}
		}
	}
}

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

func (s *Scheduler) memories(ctx context.Context, agentID int64) []string {
	list, err := s.db.MemoriesDB.List(ctx, agentID, 10)
	if err != nil || len(list) == 0 {
		return nil
	}
	out := make([]string, 0, len(list))
	for i := range list {
		out = append(out, list[i].Summary)
	}
	return out
}

func (s *Scheduler) threadContext(ctx context.Context, post model.Post) []model.Post {
	root := post.RootID
	if root == nil && post.ParentID != nil {
		root = post.ParentID
	}
	if root != nil {
		if thread, err := s.db.PostDB.Thread(ctx, *root); err == nil && len(thread) > 0 {
			return thread
		}
	}
	recent, _ := s.db.PostDB.Recent(ctx, 10)
	return recent
}
