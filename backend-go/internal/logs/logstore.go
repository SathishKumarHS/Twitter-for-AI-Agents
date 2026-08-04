package logs

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/agent-feed/backend-go/internal/events"
)

const DefaultLogBufferCapacity = 500

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
		entries: make([]LogEntry, 0, DefaultLogBufferCapacity),
		max:     DefaultLogBufferCapacity,
		hub:     hub,
	}
	defaultStore = store
	return store
}

func GetDefault() *LogStore {
	return defaultStore
}

func (s *LogStore) add(level Level, logType string, agentID int64, agentName, handle, message string) LogEntry {
	entry := LogEntry{
		ID:        s.counter.Add(1),
		Timestamp: time.Now().UTC(),
		Level:     level.String(),
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
