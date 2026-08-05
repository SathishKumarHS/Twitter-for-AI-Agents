package logs

import (
	"context"
	"log/slog"
)

func record(level Level, logType string, agentID int64, agentName, handle, message string) {
	var slogLevel slog.Level
	switch level {
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	case LevelDebug:
		slogLevel = slog.LevelDebug
	default:
		slogLevel = slog.LevelInfo
	}

	attrs := make([]slog.Attr, 0, 4)
	if logType != "" {
		attrs = append(attrs, slog.String("type", logType))
	}
	if agentID > 0 {
		attrs = append(attrs, slog.Int64("agent_id", agentID))
	}
	if agentName != "" {
		attrs = append(attrs, slog.String("agent_name", agentName))
	}
	if handle != "" {
		attrs = append(attrs, slog.String("handle", handle))
	}

	slog.LogAttrs(context.Background(), slogLevel, message, attrs...)
}

func Info(logType string, message string) {
	record(LevelInfo, logType, 0, "", "", message)
}

func Warn(logType string, message string) {
	record(LevelWarn, logType, 0, "", "", message)
}

func Error(logType string, message string) {
	record(LevelError, logType, 0, "", "", message)
}

func AgentInfo(agentID int64, agentName, handle, logType string, message string) {
	record(LevelInfo, logType, agentID, agentName, handle, message)
}

func AgentWarn(agentID int64, agentName, handle, logType string, message string) {
	record(LevelWarn, logType, agentID, agentName, handle, message)
}

func AgentError(agentID int64, agentName, handle, logType, message string) {
	record(LevelError, logType, agentID, agentName, handle, message)
}
