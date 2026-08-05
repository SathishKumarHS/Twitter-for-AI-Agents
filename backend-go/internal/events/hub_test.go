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

func TestHubMultipleSubscribers(t *testing.T) {
	hub := NewHub()
	sub1 := hub.Subscribe()
	sub2 := hub.Subscribe()
	defer hub.Unsubscribe(sub1)
	defer hub.Unsubscribe(sub2)

	ev := Event{Type: EventPostNew, Data: "broadcast post"}
	hub.Publish(ev)

	for i, sub := range []chan Event{sub1, sub2} {
		select {
		case received, ok := <-sub:
			if !ok || received.Data != "broadcast post" {
				t.Fatalf("sub %d failed to receive broadcast event", i+1)
			}
		case <-time.After(1 * time.Second):
			t.Fatalf("sub %d timed out", i+1)
		}
	}
}

func TestHubNonBlockingPublishOnFullBuffer(t *testing.T) {
	hub := NewHub()
	sub := hub.Subscribe()
	defer hub.Unsubscribe(sub)

	// Fill the buffer of size 16
	for i := 0; i < 16; i++ {
		hub.Publish(Event{Type: EventPostNew, Data: i})
	}

	// 17th publish should NOT block or panic even though buffer is full
	done := make(chan struct{})
	go func() {
		hub.Publish(Event{Type: EventPostNew, Data: "overflow"})
		close(done)
	}()

	select {
	case <-done:
		// Passed - non-blocking drop
	case <-time.After(1 * time.Second):
		t.Fatal("Publish blocked on full subscriber channel buffer")
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

func TestHubSubscribeWithNilContext(t *testing.T) {
	hub := NewHub()
	sub := hub.SubscribeWithContext(nil)
	defer hub.Unsubscribe(sub)

	if sub == nil {
		t.Fatal("expected valid subscriber channel for nil context")
	}
}
