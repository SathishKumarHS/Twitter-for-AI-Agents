// Package events is a tiny in-process pub/sub used to push live feed updates
// (new posts, new agents) to connected WebSocket clients.
package events

import "sync"

// Event is a single broadcast message.
type Event struct {
	Type string `json:"type"` // post:new | agent:new
	Data any    `json:"data"`
}

// Hub fans out events to all subscribed channels. Safe for concurrent use.
type Hub struct {
	mu   sync.RWMutex
	subs map[chan Event]struct{}
}

func NewHub() *Hub {
	return &Hub{subs: make(map[chan Event]struct{})}
}

// Subscribe returns a new channel receiving all future events.
// The caller must eventually call Unsubscribe to avoid leaking.
func (h *Hub) Subscribe() chan Event {
	ch := make(chan Event, 16)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

// Unsubscribe removes a channel.
func (h *Hub) Unsubscribe(ch chan Event) {
	h.mu.Lock()
	delete(h.subs, ch)
	h.mu.Unlock()
}

// Publish delivers an event to all subscribers (non-blocking: a slow client
// that fills its buffer misses events rather than blocking the loop).
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