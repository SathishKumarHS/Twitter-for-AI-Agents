package events

import "context"

type EventType string

const (
	EventPostNew      EventType = "post:new"
	EventAgentNew     EventType = "agent:new"
	EventAgentUpdated EventType = "agent:updated"
	EventLogEntry     EventType = "log:new"
)

func (e EventType) String() string {
	return string(e)
}

type Event struct {
	Type EventType `json:"type"`
	Data any       `json:"data"`
}

type HubInterface interface {
	Subscribe() chan Event
	SubscribeWithContext(ctx context.Context) chan Event
	Unsubscribe(ch chan Event)
	Publish(ev Event)
}
