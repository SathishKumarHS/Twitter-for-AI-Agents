// Package websocket implements the WebSocket server for streaming real-time network events to clients.
package websocket

import (
	"context"
	"net/http"
	"time"

	"nhooyr.io/websocket"

	"github.com/twitter-for-ai-agents/backend-go/internal/events"
	"github.com/twitter-for-ai-agents/backend-go/internal/server/helpers"
)

const (
	pingInterval = 20 * time.Second
	writeTimeout = 5 * time.Second
)

// WebSocket server handling client connection upgrades and event streaming
type Server struct {
	hub events.HubInterface
}

// Instantiates a new WebSocket server using the provided event hub
func New(hub events.HubInterface) *Server {
	return &Server{hub: hub}
}

// Upgrades HTTP requests to WebSockets and streams real-time events to connected clients
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	sub := s.hub.SubscribeWithContext(ctx)
	defer s.hub.Unsubscribe(sub)

	// Keepalive ping loop
	go func() {
		ticker := time.NewTicker(pingInterval)
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

	// Connection read loop to detect client disconnects
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

	// Event broadcast loop sending JSON events to the connected WebSocket client
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
			ctx2, cancel2 := context.WithTimeout(ctx, writeTimeout)
			_ = conn.Write(ctx2, websocket.MessageText, helpers.MustJSON(ev))
			cancel2()
		}
	}
}
