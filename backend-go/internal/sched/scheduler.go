package sched

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/agent-feed/backend-go/internal/db"
	"github.com/agent-feed/backend-go/internal/db/agents"
	"github.com/agent-feed/backend-go/internal/db/posts"
	"github.com/agent-feed/backend-go/internal/dispatch"
	"github.com/agent-feed/backend-go/internal/events"
)

// Scheduler is the observation loop: it wakes each active agent, feeds it new
// posts through a cheap Rule Engine gate, and dispatches relevant ones to the
// AI worker. This is the Go orchestrator's heartbeat.
type Scheduler struct {
	interval   time.Duration
	db         *db.DB
	dispatcher *dispatch.Dispatcher
	hub        *events.Hub
	cancel     context.CancelFunc
	done       chan struct{}
}

// New builds a scheduler over the given domain stores and hub.
func New(interval time.Duration, db *db.DB, d *dispatch.Dispatcher, hub *events.Hub) *Scheduler {
	return &Scheduler{
		interval:   interval,
		db:         db,
		dispatcher: d,
		hub:        hub,
		done:       make(chan struct{}),
	}
}

// Start launches the background observation loop.
func (s *Scheduler) Start(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	s.cancel = cancel
	go s.run(ctx)
	log.Printf("scheduler started (interval=%s)", s.interval)
}

// Stop terminates the loop.
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	<-s.done
}

func (s *Scheduler) run(ctx context.Context) {
	defer close(s.done)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("scheduler stopping")
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	list, err := s.db.AgentsDB.ListActive(ctx)
	if err != nil {
		log.Printf("scheduler: list agents: %v", err)
		return
	}
	for _, agent := range list {
		if err := s.processAgent(ctx, agent); err != nil {
			log.Printf("scheduler: agent %d: %v", agent.ID, err)
		}
	}
}

// processAgent does one full lifecycle turn for a single agent:
//
//	Observe → Rule Engine gate → Think (via worker) → write result
func (s *Scheduler) processAgent(ctx context.Context, agent agents.Agent) error {
	newPosts, err := s.db.PostDB.NewPostsAfter(ctx, agent.ID)
	if err != nil {
		return err
	}
	seen := s.db.PostDB.LastSeen(ctx, agent.ID)
	maxID := seen
	for _, post := range newPosts {
		if !ruleEngineRelevant(agent, post) {
			continue // skip candidates irrelevant to this agent
		}
		decision, err := s.dispatcher.Think(ctx, agent, post)
		if err != nil {
			log.Printf("scheduler: agent %d think failed: %v", agent.ID, err)
			continue
		}
		if decision.Action == "post" {
			id, err := s.db.PostDB.Create(ctx, agent.ID, decision.Text, 0)
			if err != nil {
				log.Printf("scheduler: persist post: %v", err)
				continue
			}
			s.hub.Publish(events.Event{Type: "post:new", Data: map[string]any{
				"id": id, "agent_id": agent.ID, "body": decision.Text, "kind": "post",
			}})
			_ = s.db.MemoriesDB.Add(ctx, agent.ID, "learning", "posted: "+truncate(decision.Text, 80), 5)
		} else if decision.Action == "reply" {
			id, err := s.db.PostDB.Create(ctx, agent.ID, decision.Text, post.ID)
			if err != nil {
				log.Printf("scheduler: persist reply: %v", err)
				continue
			}
			s.hub.Publish(events.Event{Type: "post:new", Data: map[string]any{
				"id": id, "agent_id": agent.ID, "body": decision.Text, "kind": "reply", "parent_id": post.ID,
			}})
			_ = s.db.MemoriesDB.Add(ctx, agent.ID, "learning", "replied to post #"+itoa(post.ID), 7)
		}
		if post.ID > maxID {
			maxID = post.ID
		}
	}
	return s.db.PostDB.SetSeen(ctx, agent.ID, maxID)
}

// ruleEngineRelevant is the cheap, deterministic pre-filter run in Go so the
// LLM worker only fires for topics the agent actually cares about.
func ruleEngineRelevant(agent agents.Agent, post posts.Post) bool {
	for _, id := range post.Mentions {
		if id == agent.ID {
			return true // direct mention → always relevant
		}
	}
	for _, kw := range agent.Interests {
		if kw != "" && strings.Contains(strings.ToLower(post.Body), strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

func itoa(v int64) string {
	return strconv.FormatInt(v, 10)
}