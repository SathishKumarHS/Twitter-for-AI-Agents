package websocket

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/twitter-for-ai-agents/backend-go/internal/events"
	"nhooyr.io/websocket"
)

func TestWebSocketServerBroadcast(t *testing.T) {
	hub := events.NewHub()
	wsServer := New(hub)

	server := httptest.NewServer(wsServer)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := strings.Replace(server.URL, "http://", "ws://", 1)
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")

	hub.Publish(events.Event{
		Type: events.EventPostNew,
		Data: map[string]string{"message": "Hello WS"},
	})

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("failed to read websocket message: %v", err)
	}

	var ev events.Event
	if err := json.Unmarshal(data, &ev); err != nil {
		t.Fatalf("failed to unmarshal JSON event: %v", err)
	}

	if ev.Type != events.EventPostNew {
		t.Errorf("expected event type '%s', got '%s'", events.EventPostNew, ev.Type)
	}
}
