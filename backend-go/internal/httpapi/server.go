package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"nhooyr.io/websocket"

	"github.com/agent-feed/backend-go/internal/cache"
	"github.com/agent-feed/backend-go/internal/db"
	"github.com/agent-feed/backend-go/internal/events"
)

// Server serves the REST + WebSocket API for the frontend.
type Server struct {
	addr  string
	db    *db.DB
	hub   *events.Hub
	cache cache.CacheInterface
	srv   *http.Server
}

// New returns a configured HTTP server (not yet listening).
func New(addr string, db *db.DB, hub *events.Hub, c cache.CacheInterface) *Server {
	return &Server{addr: addr, db: db, hub: hub, cache: c}
}

// ListenAndServe binds the server and blocks until stopped.
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	s.routes(mux)
	s.srv = &http.Server{
		Addr:              s.addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s.srv.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.srv == nil {
		return nil
	}
	return s.srv.Shutdown(ctx)
}

func (s *Server) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	// REST
	mux.HandleFunc("GET /api/agents", s.handleListAgents)
	mux.HandleFunc("POST /api/agents", s.handleCreateAgent)
	mux.HandleFunc("GET /api/agents/{id}", s.handleGetAgent)
	mux.HandleFunc("POST /api/agents/{id}/pause", s.handlePauseAgent)
	mux.HandleFunc("POST /api/agents/{id}/resume", s.handleResumeAgent)
	mux.HandleFunc("GET /api/agents/{id}/posts", s.handleAgentPosts)
	mux.HandleFunc("GET /api/posts", s.handleFeed)
	mux.HandleFunc("GET /api/posts/thread/{id}", s.handleThread)
	mux.HandleFunc("GET /api/logs", s.handleGetLogs)

	// WebSocket
	mux.HandleFunc("/ws", s.handleWebSocket)
}

// handleWebSocket upgrades the connection and streams hub events to the client.
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusInternalError, "closed")

	if s.hub == nil {
		return
	}
	sub := s.hub.Subscribe()
	defer s.hub.Unsubscribe(sub)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Ping keeps the connection alive; reads also surface client disconnects.
	go func() {
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := conn.Ping(ctx); err != nil {
					cancel()
					return
				}
			}
		}
	}()

	// Register a close-notification reader so we stop when the client goes away.
	closeCh := make(chan struct{})
	go func() {
		defer close(closeCh)
		for {
			if _, _, err := conn.Read(ctx); err != nil {
				cancel()
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-closeCh:
			return
		case ev, ok := <-sub:
			if !ok {
				return
			}
			ctx2, cancel2 := context.WithTimeout(ctx, 5*time.Second)
			_ = conn.Write(ctx2, websocket.MessageText, mustJSON(ev))
			cancel2()
		}
	}
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return b
}
