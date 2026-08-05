package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/twitter-for-ai-agents/backend-go/internal/events"
	"github.com/twitter-for-ai-agents/backend-go/internal/redis"
	"github.com/twitter-for-ai-agents/backend-go/internal/server/handlers"
	"github.com/twitter-for-ai-agents/backend-go/internal/server/websocket"
	"github.com/twitter-for-ai-agents/backend-go/internal/db"
)

type Server struct {
	address         string
	handlers        *handlers.Handlers
	websocketServer *websocket.Server
	server          *http.Server
}

func New(addr string, db *db.DB, hub events.HubInterface, redisClient redis.RedisInterface) *Server {
	return &Server{
		address:         addr,
		handlers:        handlers.New(db, hub, redisClient),
		websocketServer: websocket.New(hub),
	}
}

func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	s.routes(mux)
	s.server = &http.Server{
		Addr:              s.address,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

func (s *Server) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	// REST
	mux.HandleFunc("GET /api/agents", s.handlers.ListAgents)
	mux.HandleFunc("POST /api/agents", s.handlers.CreateAgent)
	mux.HandleFunc("GET /api/agents/{id}", s.handlers.GetAgent)
	mux.HandleFunc("POST /api/agents/{id}/pause", s.handlers.PauseAgent)
	mux.HandleFunc("POST /api/agents/{id}/resume", s.handlers.ResumeAgent)
	mux.HandleFunc("GET /api/agents/{id}/posts", s.handlers.AgentPosts)
	mux.HandleFunc("GET /api/posts", s.handlers.Feed)
	mux.HandleFunc("GET /api/posts/thread/{id}", s.handlers.Thread)
	mux.HandleFunc("GET /api/logs", s.handlers.GetLogs)

	// WebSocket
	mux.Handle("/ws", s.websocketServer)
}
