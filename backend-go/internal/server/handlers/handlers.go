// Package handlers implements HTTP REST API request handlers for agents and posts.
package handlers

import (
	"github.com/twitter-for-ai-agents/backend-go/internal/db"
	"github.com/twitter-for-ai-agents/backend-go/internal/events"
	"github.com/twitter-for-ai-agents/backend-go/internal/redis"
)

// Container for HTTP REST handlers sharing database, event hub, and redis cache
type Handlers struct {
	db          *db.DB
	hub         events.HubInterface
	redisClient redis.RedisInterface
}

// Instantiates Handlers with database, event hub, and redis dependencies
func New(db *db.DB, hub events.HubInterface, redisClient redis.RedisInterface) *Handlers {
	return &Handlers{
		db:          db,
		hub:         hub,
		redisClient: redisClient,
	}
}

// Helper: publishes an event to the real-time event hub if configured
func (h *Handlers) publish(ev events.Event) {
	if h.hub != nil {
		h.hub.Publish(ev)
	}
}
