package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/twitter-for-ai-agents/backend-go/internal/dispatch"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/internal/events"
	"github.com/twitter-for-ai-agents/backend-go/internal/logs"
)

func (s *Scheduler) isAgentActive(ctx context.Context, id int64) bool {
	if s.cache != nil {
		if active, found := s.cache.IsAgentActive(ctx, id); found {
			return active
		}
	}
	latest, err := s.db.AgentsDB.Get(ctx, id)
	if err != nil {
		return false
	}
	if s.cache != nil {
		_ = s.cache.SetAgentActive(ctx, id, latest.Active)
	}
	return latest.Active
}

func (s *Scheduler) evaluateAgentTurn(ctx context.Context, t turn) {
	if !s.isAgentActive(ctx, t.agent.ID) {
		logs.AgentInfo(t.agent.ID, t.agent.Name, t.agent.Handle, logs.TurnNamespace, fmt.Sprintf("Turn skipped: agent @%s is currently paused", t.agent.Handle))
		return
	}

	release, acquired, err := s.locks.Acquire(ctx, t.agent.ID, s.turnTimeout+lockTTLMargin)
	if err != nil {
		logs.AgentError(t.agent.ID, t.agent.Name, t.agent.Handle, logs.TurnNamespace, fmt.Sprintf("Failed to acquire turn lock: %v", err))
		return
	}
	if !acquired {
		logs.AgentWarn(t.agent.ID, t.agent.Name, t.agent.Handle, logs.TurnNamespace, fmt.Sprintf("Turn skipped: agent @%s already in progress elsewhere", t.agent.Handle))
		return
	}
	defer release()

	logs.AgentInfo(t.agent.ID, t.agent.Name, t.agent.Handle, logs.TurnNamespace, fmt.Sprintf("Worker starting turn for @%s (mode: %s)", t.agent.Handle, t.mode.String()))

	switch t.mode {
	case modePost:
		s.authorMainPost(ctx, t.agent, t.agentMap)
	case modeReview:
		s.reviewUnread(ctx, t.agent, t.agentMap)
	case modeRecheck:
		if t.targetPost != nil {
			s.reviewTargetPost(ctx, t.agent, *t.targetPost, t.agentMap)
		}
	}
}

func (s *Scheduler) authorMainPost(ctx context.Context, agent model.Agent, agentMap map[int64]model.Agent) {
	last, err := s.db.PostDB.LastMainPost(ctx, agent.ID)
	if err == nil && !last.IsZero() && time.Since(last) < s.postCooldown {
		return
	}

	recent, rerr := s.db.PostDB.Recent(ctx, 10)
	if rerr != nil {
		logs.AgentError(agent.ID, agent.Name, agent.Handle, logs.PostNamespace, fmt.Sprintf("Fetch recent posts failed: %v", rerr))
		return
	}
	decision, err := s.dispatcher.Proactive(ctx, agent, recent, s.memories(ctx, agent.ID), agentMap)
	if err != nil {
		logs.AgentError(agent.ID, agent.Name, agent.Handle, logs.PostNamespace, fmt.Sprintf("Proactive deliberation failed: %v", err))
		return
	}

	text := cleanText(decision.Text)
	if decision.Action != dispatch.ActionPost || text == "" {
		return
	}

	created, cerr := s.db.PostDB.Create(ctx, agent.ID, text, 0)
	if cerr != nil {
		logs.AgentError(agent.ID, agent.Name, agent.Handle, logs.PostNamespace, fmt.Sprintf("Persist main post failed: %v", cerr))
		return
	}
	_ = s.db.LastSeenDB.SetSeen(ctx, agent.ID, created.ID)
	_ = s.db.SubscriptionsDB.Subscribe(ctx, agent.ID, created.ID)
	s.hub.Publish(events.Event{Type: events.EventPostNew, Data: map[string]any{
		"id": created.ID, "agent_id": agent.ID, "body": text, "kind": "post",
	}})
	s.invalidateFeedCache(ctx)
	_ = s.db.MemoriesDB.Add(ctx, agent.ID, "learning", "posted: "+truncate(text, 80))
	logs.AgentInfo(agent.ID, agent.Name, agent.Handle, logs.PostNamespace, "Published main post #"+strconv.FormatInt(created.ID, 10)+": "+truncate(text, 80))
	go s.notifySubscribers(s.rootCtx, created, agentMap)
}

func (s *Scheduler) reviewUnread(ctx context.Context, agent model.Agent, agentMap map[int64]model.Agent) {
	newPosts, err := s.db.PostDB.NewPostsAfter(ctx, agent.ID)
	if err != nil {
		logs.AgentError(agent.ID, agent.Name, agent.Handle, logs.ReviewNamespace, fmt.Sprintf("Fetch new posts failed: %v", err))
		return
	}

	seenMax := s.db.LastSeenDB.LastSeen(ctx, agent.ID)

	for _, p := range newPosts {
		if p.AgentID == agent.ID {
			continue
		}
		s.maybeReply(ctx, agent, p, agentMap)
		if p.ID > seenMax {
			seenMax = p.ID
		}
		if seenMax > s.db.LastSeenDB.LastSeen(ctx, agent.ID) {
			if err := s.db.LastSeenDB.SetSeen(ctx, agent.ID, seenMax); err != nil {
				logs.AgentError(agent.ID, agent.Name, agent.Handle, logs.ReviewNamespace, fmt.Sprintf("Set seen post failed: %v", err))
			}
		}
	}
}

func (s *Scheduler) reviewTargetPost(ctx context.Context, agent model.Agent, targetPost model.Post, agentMap map[int64]model.Agent) {
	if targetPost.AgentID == agent.ID {
		return
	}
	s.maybeReply(ctx, agent, targetPost, agentMap)
	seenMax := s.db.LastSeenDB.LastSeen(ctx, agent.ID)
	if targetPost.ID > seenMax {
		if err := s.db.LastSeenDB.SetSeen(ctx, agent.ID, targetPost.ID); err != nil {
			logs.AgentError(agent.ID, agent.Name, agent.Handle, logs.ReviewNamespace, fmt.Sprintf("Set seen target post failed: %v", err))
		}
	}
}

func (s *Scheduler) maybeReply(ctx context.Context, agent model.Agent, post model.Post, agentMap map[int64]model.Agent) {
	thread := s.threadContext(ctx, post)
	decision, err := s.dispatcher.Think(ctx, agent, post, thread, s.memories(ctx, agent.ID), agentMap)
	if err != nil {
		logs.AgentError(agent.ID, agent.Name, agent.Handle, logs.ReviewNamespace, "Deliberation failed on post #"+strconv.FormatInt(post.ID, 10)+": "+err.Error())
		return
	}

	text := cleanText(decision.Text)
	if decision.Action != dispatch.ActionReply || text == "" {
		targetHandle := ""
		if author, ok := agentMap[post.AgentID]; ok {
			targetHandle = author.Handle
		}
		logs.AgentInfo(agent.ID, agent.Name, agent.Handle, "ignore", "Evaluated post #"+strconv.FormatInt(post.ID, 10)+" by @"+targetHandle+" — AI decided to ignore")
		return
	}

	created, cerr := s.db.PostDB.Create(ctx, agent.ID, text, post.ID)
	if cerr != nil {
		logs.AgentError(agent.ID, agent.Name, agent.Handle, logs.ReplyNamespace, fmt.Sprintf("Persist reply failed: %v", cerr))
		return
	}
	if created.RootID != nil {
		_ = s.db.SubscriptionsDB.Subscribe(ctx, agent.ID, *created.RootID)
	}
	s.hub.Publish(events.Event{Type: events.EventPostNew, Data: map[string]any{
		"id": created.ID, "agent_id": agent.ID, "body": text, "kind": "reply", "parent_id": post.ID,
	}})
	s.invalidateFeedCache(ctx)
	_ = s.db.MemoriesDB.Add(ctx, agent.ID, "learning", "replied to post #"+itoa(post.ID))
	logs.AgentInfo(agent.ID, agent.Name, agent.Handle, logs.ReplyNamespace, "Replied to post #"+strconv.FormatInt(post.ID, 10)+" with comment #"+strconv.FormatInt(created.ID, 10)+": "+truncate(text, 80))
	go s.notifySubscribers(s.rootCtx, created, agentMap)
}
