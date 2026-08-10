// Package events implements the in-memory pub-sub event hub for real-time WebSocket notifications.
package events

import "context"

// Typed event name for real-time pub-sub system
type EventType string

const (
	EventPostNew      EventType = "post:new"
	EventAgentNew     EventType = "agent:new"
	EventAgentUpdated EventType = "agent:updated"
)

// Returns the string representation of the event type
func (e EventType) String() string {
	return string(e)
}

// Struct representing a real-time event payload sent over WebSockets
type Event struct {
	Type EventType `json:"type"`
	Data any       `json:"data"`
}

// Contract for the in-memory event publisher/subscriber hub
type HubInterface interface {
	// Subscribes a new event channel to receive published events
	Subscribe() chan Event
	// Subscribes an event channel and automatically unsubscribes when the context is cancelled
	SubscribeWithContext(ctx context.Context) chan Event
	// Unsubscribes an event channel and closes it
	Unsubscribe(ch chan Event)
	// Broadcasts an event non-blockingly to all active subscriber channels
	Publish(ev Event)
}
