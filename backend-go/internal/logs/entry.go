package logs

import "time"

type Level string

const (
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
	LevelDebug Level = "debug"
)

const (
	SystemNamespace = "system"
	TurnNamespace   = "turn"
	PostNamespace   = "post"
	ReviewNamespace = "review"
	ReplyNamespace  = "reply"
)

func (l Level) String() string {
	return string(l)
}

type LogEntry struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	AgentID   int64     `json:"agent_id,omitempty"`
	AgentName string    `json:"agent_name,omitempty"`
	Handle    string    `json:"handle,omitempty"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
}
