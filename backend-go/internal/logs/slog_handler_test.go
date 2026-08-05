package logs

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/twitter-for-ai-agents/backend-go/internal/events"
)

func TestHubHandlerCapturesSlog(t *testing.T) {
	hub := events.NewHub()
	store := Init(hub)

	sub := hub.Subscribe()
	defer hub.Unsubscribe(sub)

	buf := &bytes.Buffer{}
	handler := NewHubHandler(buf, store, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(handler)

	logger.Info("Test structured log",
		slog.String("type", SystemNamespace),
		slog.Int64("agent_id", 42),
		slog.String("handle", "agent_bot"),
	)

	// Check stdout output
	out := buf.String()
	if out == "" {
		t.Fatal("expected non-empty output in writer")
	}

	// Check LogStore buffer
	recent := store.Recent(1)
	if len(recent) == 0 {
		t.Fatal("expected entry in LogStore")
	}
	if recent[0].AgentID != 42 || recent[0].Handle != "agent_bot" || recent[0].Message != "Test structured log" {
		t.Fatalf("unexpected LogStore entry: %+v", recent[0])
	}

	// Check event hub broadcast
	select {
	case ev := <-sub:
		if ev.Type != events.EventLogEntry {
			t.Fatalf("expected log:new event type, got %s", ev.Type)
		}
	default:
		t.Fatal("expected event in hub subscriber channel")
	}
}
