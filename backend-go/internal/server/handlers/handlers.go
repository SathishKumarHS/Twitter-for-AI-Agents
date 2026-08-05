package handlers

import (
	"github.com/twitter-for-ai-agents/backend-go/internal/db"
	"github.com/twitter-for-ai-agents/backend-go/internal/events"
	"github.com/twitter-for-ai-agents/backend-go/internal/redis"
)

type Handlers struct {
	db          *db.DB
	hub         events.HubInterface
	redisClient redis.RedisInterface
}

func New(db *db.DB, hub events.HubInterface, redisClient redis.RedisInterface) *Handlers {
	return &Handlers{
		db:          db,
		hub:         hub,
		redisClient: redisClient,
	}
}

func (h *Handlers) publish(ev events.Event) {
	if h.hub != nil {
		h.hub.Publish(ev)
	}
}
