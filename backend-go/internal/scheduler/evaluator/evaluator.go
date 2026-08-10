// Package evaluator executes single agent deliberation turns with turn locking and mode routing.
package evaluator

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/twitter-for-ai-agents/backend-go/internal/db"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/internal/dispatch"
	"github.com/twitter-for-ai-agents/backend-go/internal/redis"
)

const lockTTLMargin = 60 * time.Second
const maxReviewBatch = 10

// Internal scheduler turn mode classification
type TurnMode string

const (
	ModePost    TurnMode = "author_main_post"
	ModeReview  TurnMode = "review_unread_feed"
	ModeRecheck TurnMode = "comment_recheck"
)

// Returns string representation of TurnMode
func (m TurnMode) String() string {
	return string(m)
}

// Holds turn context data submitted for evaluation
type Turn struct {
	Mode       TurnMode
	Agent      model.Agent
	AgentMap   map[int64]model.Agent
	TargetPost *model.Post
}

// Interface contract for turn evaluator
type EvaluatorInterface interface {
	EvaluateAgentTurn(ctx context.Context, t Turn)
	IsAgentActive(ctx context.Context, id int64) bool
}

// Handles lock acquisition, unread post fetching, and turn dispatching
type Evaluator struct {
	db          *db.DB
	dispatcher  dispatch.DispatcherInterface
	locks       redis.RedisInterface
	cache       redis.RedisInterface
	turnTimeout time.Duration
}

// Instantiates a new Evaluator with database, dispatcher, redis locks, and turn timeout
func New(
	database *db.DB,
	dispatcher dispatch.DispatcherInterface,
	locks redis.RedisInterface,
	cache redis.RedisInterface,
	turnTimeout time.Duration,
) *Evaluator {
	return &Evaluator{
		db:           database,
		dispatcher:   dispatcher,
		locks:        locks,
		cache:        cache,
		turnTimeout:  turnTimeout,
	}
}

// Checks if an agent is active (using Redis cache first, falling back to database)
func (e *Evaluator) IsAgentActive(ctx context.Context, id int64) bool {
	if e.cache != nil {
		if active, found := e.cache.IsAgentActive(ctx, id); found {
			return active
		}
	}
	latest, err := e.db.AgentsDB.Get(ctx, id)
	if err != nil {
		return false
	}
	if e.cache != nil {
		_ = e.cache.SetAgentActive(ctx, id, latest.Active)
	}
	return latest.Active
}

// Executes a single agent deliberation turn synchronously after acquiring a turn lock
func (e *Evaluator) EvaluateAgentTurn(ctx context.Context, t Turn) {
	if !e.IsAgentActive(ctx, t.Agent.ID) {
		slog.Info("Turn skipped", "mode", t.Mode, "agent_id", t.Agent.ID, "agent_name", t.Agent.Name, "agent_handle", t.Agent.Handle, "reason", "paused")
		return
	}

	release, acquired, err := e.locks.Acquire(ctx, t.Agent.ID, e.turnTimeout+lockTTLMargin)
	if err != nil {
		slog.Error("Failed to acquire turn lock", "error", err, "agent_id", t.Agent.ID, "agent_name", t.Agent.Name, "agent_handle", t.Agent.Handle)
		return
	}
	if !acquired {
		slog.Info("Turn skipped", "mode", t.Mode, "agent_id", t.Agent.ID, "agent_name", t.Agent.Name, "agent_handle", t.Agent.Handle, "reason", "already in progress")
		return
	}
	defer release()

	slog.Info("Turn started", "mode", t.Mode, "agent_id", t.Agent.ID, "agent_name", t.Agent.Name, "agent_handle", t.Agent.Handle, "timeout", e.turnTimeout)

	turnCtx, cancel := context.WithTimeout(ctx, e.turnTimeout)
	defer cancel()

	switch t.Mode {
	case ModePost:
		e.authorMainPost(turnCtx, t.Agent)
	case ModeReview:
		e.reviewUnread(turnCtx, t.Agent, t.AgentMap)
	case ModeRecheck:
		if t.TargetPost != nil {
			e.reviewTargetPost(turnCtx, t.Agent, *t.TargetPost, t.AgentMap)
		}
	}
}

// Dispatches a proactive main post turn to the AI Worker
func (e *Evaluator) authorMainPost(ctx context.Context, agent model.Agent) {
	done := make(chan struct{})
	e.dispatcher.Act(ctx, dispatch.ModeProactive, agent, model.Post{}, func(decision *dispatch.Decision, err error) {
		defer close(done)
		if err != nil {
			slog.Error("Proactive turn failed", "error", err, "agent_id", agent.ID, "agent_name", agent.Name, "agent_handle", agent.Handle)
			return
		}
		slog.Info("Proactive turn finished", "executed", decision.Executed, "agent_id", agent.ID, "agent_name", agent.Name, "agent_handle", agent.Handle)
	})
	<-done
}

// Fetches unread posts for an agent and dispatches review turns in batches
func (e *Evaluator) reviewUnread(ctx context.Context, agent model.Agent, agentMap map[int64]model.Agent) {
	newPosts, err := e.db.PostDB.NewPostsAfter(ctx, agent.ID)
	if err != nil {
		slog.Error("Fetch new posts failed", "error", err, "agent_id", agent.ID, "agent_name", agent.Name, "agent_handle", agent.Handle)
		return
	}

	if len(newPosts) == 0 {
		return
	}

	maxID := newPosts[len(newPosts)-1].ID
	defer func() {
		_ = e.db.LastSeenDB.SetSeen(ctx, agent.ID, maxID)
	}()

	selected := filterPostsForAgent(newPosts, agent.ID)
	if len(selected) > maxReviewBatch {
		selected = selected[:maxReviewBatch]
	}

	var wg sync.WaitGroup
	for _, post := range selected {
		wg.Add(1)
		p := post
		e.dispatcher.Act(ctx, dispatch.ModeRespond, agent, p, func(decision *dispatch.Decision, err error) {
			defer wg.Done()
			if err != nil {
				slog.Error("Review turn failed", "error", err, "agent_id", agent.ID, "agent_name", agent.Name, "agent_handle", agent.Handle, "post_id", p.ID)
				return
			}
			slog.Info("Review turn finished", "executed", decision.Executed, "agent_id", agent.ID, "agent_name", agent.Name, "agent_handle", agent.Handle, "post_id", p.ID)
		})
	}
	wg.Wait()
}

// Dispatches a reactive turn for a specific target post
func (e *Evaluator) reviewTargetPost(ctx context.Context, agent model.Agent, post model.Post, agentMap map[int64]model.Agent) {
	done := make(chan struct{})
	e.dispatcher.Act(ctx, dispatch.ModeRespond, agent, post, func(decision *dispatch.Decision, err error) {
		defer close(done)
		if err != nil {
			slog.Error("Recheck turn failed", "error", err, "agent_id", agent.ID, "agent_name", agent.Name, "agent_handle", agent.Handle, "post_id", post.ID)
			return
		}
		slog.Info("Recheck turn finished", "executed", decision.Executed, "agent_id", agent.ID, "agent_name", agent.Name, "agent_handle", agent.Handle, "post_id", post.ID)
	})
	<-done
}

// Filters candidate unread posts prioritizing direct mentions and relevant posts
func filterPostsForAgent(posts []model.Post, agentID int64) []model.Post {
	var mentioned []model.Post
	var general []model.Post

	for _, p := range posts {
		if p.AgentID == agentID {
			continue
		}
		isMentioned := false
		for _, m := range p.Mentions {
			if m == agentID {
				isMentioned = true
				break
			}
		}
		if isMentioned {
			mentioned = append(mentioned, p)
		} else {
			general = append(general, p)
		}
	}
	return append(mentioned, general...)
}
