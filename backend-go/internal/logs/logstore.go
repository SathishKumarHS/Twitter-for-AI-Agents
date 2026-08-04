package logs

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/agent-feed/backend-go/internal/events"
)

type LogEntry struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"` // "info" | "warn" | "error" | "debug"
	AgentID   int64     `json:"agent_id,omitempty"`
	AgentName string    `json:"agent_name,omitempty"`
	Handle    string    `json:"handle,omitempty"`
	Message   string    `json:"message"`
	Type      string    `json:"type"` // "post" | "reply" | "cooldown" | "review" | "system"
}

// LogStore holds a ring buffer of recent system logs and streams them to the UI hub.
type LogStore struct {
	mu      sync.RWMutex
	entries []LogEntry
	max     int
	hub     events.HubInterface
	counter atomic.Int64
}

// Global default log store reference
var defaultStore *LogStore

func Init(hub events.HubInterface) *LogStore {
	store := &LogStore{
		entries: make([]LogEntry, 0, 500),
		max:     500,
		hub:     hub,
	}
	defaultStore = store
	return store
}

func GetDefault() *LogStore {
	return defaultStore
}

func (s *LogStore) Add(level, logType string, agentID int64, agentName, handle, message string) LogEntry {
	entry := LogEntry{
		ID:        s.counter.Add(1),
		Timestamp: time.Now().UTC(),
		Level:     level,
		Type:      logType,
		AgentID:   agentID,
		AgentName: agentName,
		Handle:    handle,
		Message:   message,
	}

	s.mu.Lock()
	if len(s.entries) >= s.max {
		s.entries = s.entries[1:]
	}
	s.entries = append(s.entries, entry)
	s.mu.Unlock()

	if s.hub != nil {
		s.hub.Publish(events.Event{
			Type: events.EventLogEntry,
			Data: entry,
		})
	}

	return entry
}

func (s *LogStore) Recent(limit int) []LogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.entries) {
		limit = len(s.entries)
	}

	result := make([]LogEntry, limit)
	start := len(s.entries) - limit
	copy(result, s.entries[start:])
	return result
}

// Global convenience functions
func Record(level, logType string, agentID int64, agentName, handle, message string) {
	if defaultStore != nil {
		defaultStore.Add(level, logType, agentID, agentName, handle, message)
	}
}

func Info(logType, message string) {
	Record("info", logType, 0, "", "", message)
}

func AgentInfo(agentID int64, agentName, handle, logType, message string) {
	Record("info", logType, agentID, agentName, handle, message)
}

func AgentWarn(agentID int64, agentName, handle, logType, message string) {
	Record("warn", logType, agentID, agentName, handle, message)
}

func AgentError(agentID int64, agentName, handle, logType, message string) {
	Record("error", logType, agentID, agentName, handle, message)
}
