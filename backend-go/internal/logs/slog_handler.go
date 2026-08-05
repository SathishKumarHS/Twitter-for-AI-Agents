package logs

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// HubHandler is a custom slog.Handler that outputs structured logs to stdout
// and automatically forwards entries to the LogStore and events.Hub.
type HubHandler struct {
	slog.Handler
	store *LogStore
}

func NewHubHandler(w io.Writer, store *LogStore, opts *slog.HandlerOptions) *HubHandler {
	if w == nil {
		w = os.Stdout
	}
	if opts == nil {
		opts = &slog.HandlerOptions{Level: slog.LevelInfo}
	}
	base := slog.NewTextHandler(w, opts)
	return &HubHandler{
		Handler: base,
		store:   store,
	}
}

func (h *HubHandler) Handle(ctx context.Context, r slog.Record) error {
	// First write to stdout via base handler
	err := h.Handler.Handle(ctx, r)

	// Forward to LogStore (which broadcasts to events.Hub)
	if h.store != nil {
		var agentID int64
		var agentName, handle, logType string

		r.Attrs(func(a slog.Attr) bool {
			switch a.Key {
			case "agent_id":
				agentID = a.Value.Int64()
			case "agent_name":
				agentName = a.Value.String()
			case "handle":
				handle = a.Value.String()
			case "type", "namespace":
				logType = a.Value.String()
			}
			return true
		})

		if logType == "" {
			logType = SystemNamespace
		}

		level := LevelInfo
		switch r.Level {
		case slog.LevelWarn:
			level = LevelWarn
		case slog.LevelError:
			level = LevelError
		case slog.LevelDebug:
			level = LevelDebug
		}

		h.store.add(level, logType, agentID, agentName, handle, r.Message)
	}

	return err
}
