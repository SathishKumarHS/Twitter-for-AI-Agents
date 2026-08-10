package events

import (
	"context"
	"sync"
)

// Thread-safe in-memory publish-subscribe event hub implementation
type Hub struct {
	mu   sync.RWMutex
	subs map[chan Event]struct{}
}

// Instantiates a new in-memory event Hub
func NewHub() HubInterface {
	return &Hub{subs: make(map[chan Event]struct{})}
}

// Registers and returns a new buffered event channel for subscribing to published events
func (h *Hub) Subscribe() chan Event {
	ch := make(chan Event, 16)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

// Subscribes to events and launches a background goroutine to clean up when context finishes
func (h *Hub) SubscribeWithContext(ctx context.Context) chan Event {
	ch := h.Subscribe()
	if ctx == nil {
		return ch
	}
	go func() {
		<-ctx.Done()
		h.Unsubscribe(ch)
	}()
	return ch
}

// Removes the channel from active subscribers and closes it safely
func (h *Hub) Unsubscribe(ch chan Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.subs[ch]; ok {
		delete(h.subs, ch)
		close(ch)
	}
}

// Broadcasts an event to all subscribers without blocking slow receivers
func (h *Hub) Publish(ev Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subs {
		select {
		case ch <- ev:
		default:
		}
	}
}
