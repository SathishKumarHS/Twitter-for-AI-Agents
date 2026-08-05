package events

import (
	"context"
	"testing"
	"time"
)

func TestHubSubscribeAndPublish(t *testing.T) {
	hub := NewHub()
	sub := hub.Subscribe()
	defer hub.Unsubscribe(sub)

	ev := Event{Type: EventPostNew, Data: "test post"}
	hub.Publish(ev)

	select {
	case received, ok := <-sub:
		if !ok {
			t.Fatal("expected open channel, got closed")
		}
		if received.Type != EventPostNew || received.Data != "test post" {
			t.Fatalf("unexpected event: %+v", received)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for published event")
	}
}

func TestHubUnsubscribeClosesChannel(t *testing.T) {
	hub := NewHub()
	sub := hub.Subscribe()
	hub.Unsubscribe(sub)

	select {
	case _, ok := <-sub:
		if ok {
			t.Fatal("expected channel to be closed after Unsubscribe")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for closed channel")
	}

	// Idempotent call to Unsubscribe should not panic
	hub.Unsubscribe(sub)
}

func TestHubSubscribeWithContextAutoCleanup(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())

	sub := hub.SubscribeWithContext(ctx)

	// Cancel context to trigger auto-unsubscribe
	cancel()

	select {
	case _, ok := <-sub:
		if ok {
			t.Fatal("expected channel to be closed automatically after context cancellation")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for context cancellation cleanup")
	}
}
