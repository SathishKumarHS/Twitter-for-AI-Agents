// Package handlers implements HTTP REST API request handlers for agents and posts.
package handlers

import (
	"context"

	"github.com/twitter-for-ai-agents/backend-go/internal/db"
	"github.com/twitter-for-ai-agents/backend-go/internal/events"
	"github.com/twitter-for-ai-agents/backend-go/internal/redis"
	"github.com/twitter-for-ai-agents/backend-go/internal/scheduler/evaluator"
)

// EvaluatorInterface abstracts turn evaluation for handlers
type EvaluatorInterface interface {
	EvaluateAgentTurn(ctx context.Context, t evaluator.Turn)
}

// Container for HTTP REST handlers sharing database, event hub, redis cache, and turn evaluator
type Handlers struct {
	db          *db.DB
	hub         events.HubInterface
	redisClient redis.RedisInterface
	evaluator   EvaluatorInterface
}

// Instantiates Handlers with database, event hub, redis, and evaluator dependencies
func New(db *db.DB, hub events.HubInterface, redisClient redis.RedisInterface, eval EvaluatorInterface) *Handlers {
	return &Handlers{
		db:          db,
		hub:         hub,
		redisClient: redisClient,
		evaluator:   eval,
	}
}

// Helper: publishes an event to the real-time event hub if configured
func (h *Handlers) publish(ev events.Event) {
	if h.hub != nil {
		h.hub.Publish(ev)
	}
}
