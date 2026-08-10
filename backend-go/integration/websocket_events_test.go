//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"

	"github.com/twitter-for-ai-agents/backend-go/internal/events"
)

func connectWebSocket(t *testing.T, ctx context.Context) (*websocket.Conn, context.Context) {
	t.Helper()
	wsURL := strings.Replace(baseURL, "http://", "ws://", 1) + "/ws"
	wsCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	conn, _, err := websocket.Dial(wsCtx, wsURL, nil)
	if err != nil {
		cancel()
		t.Fatalf("dial websocket: %v", err)
	}
	t.Cleanup(func() {
		conn.Close(websocket.StatusNormalClosure, "done")
		cancel()
	})
	time.Sleep(100 * time.Millisecond)
	return conn, wsCtx
}

func readWebsocketEvent(t *testing.T, conn *websocket.Conn, ctx context.Context, want events.EventType) {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		_, data, err := conn.Read(ctx)
		if err != nil {
			t.Fatalf("read websocket: %v", err)
		}
		var ev events.Event
		if err := json.Unmarshal(data, &ev); err != nil {
			t.Fatalf("unmarshal event: %v", err)
		}
		if ev.Type == want {
			return
		}
	}
	t.Fatalf("did not receive %s event on websocket", want)
}

func TestWebSocketReceivesLiveEvents(t *testing.T) {
	ctx := context.Background()
	reset(ctx)

	conn, wsCtx := connectWebSocket(t, ctx)
	createAgent(t, "ws_agent")

	readWebsocketEvent(t, conn, wsCtx, events.EventAgentNew)
}
