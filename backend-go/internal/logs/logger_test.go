package logs

import (
	"testing"

	"github.com/twitter-for-ai-agents/backend-go/internal/events"
)

func TestLoggerHelpers(t *testing.T) {
	hub := events.NewHub()
	store := Init(hub)

	Info(SystemNamespace, "System starting")
	Warn(SystemNamespace, "Low disk space warning")
	Error(SystemNamespace, "Database connection failed")

	AgentInfo(10, "BotOne", "bot_one", TurnNamespace, "Processing turn")
	AgentWarn(10, "BotOne", "bot_one", TurnNamespace, "Retry turn")
	AgentError(10, "BotOne", "bot_one", TurnNamespace, "Turn crashed")

	recent := store.Recent(6)
	if len(recent) != 6 {
		t.Fatalf("expected 6 log entries, got %d", len(recent))
	}

	if recent[0].Level != string(LevelInfo) || recent[0].Message != "System starting" {
		t.Errorf("unexpected log 0: %+v", recent[0])
	}
	if recent[1].Level != string(LevelWarn) || recent[1].Message != "Low disk space warning" {
		t.Errorf("unexpected log 1: %+v", recent[1])
	}
	if recent[2].Level != string(LevelError) || recent[2].Message != "Database connection failed" {
		t.Errorf("unexpected log 2: %+v", recent[2])
	}

	if recent[3].AgentID != 10 || recent[3].Handle != "bot_one" || recent[3].Message != "Processing turn" {
		t.Errorf("unexpected agent log 3: %+v", recent[3])
	}
	if recent[4].AgentID != 10 || recent[4].Level != string(LevelWarn) {
		t.Errorf("unexpected agent log 4: %+v", recent[4])
	}
	if recent[5].AgentID != 10 || recent[5].Level != string(LevelError) {
		t.Errorf("unexpected agent log 5: %+v", recent[5])
	}
}

func TestLevelString(t *testing.T) {
	if LevelInfo.String() != "info" {
		t.Errorf("expected 'info', got '%s'", LevelInfo.String())
	}
	if LevelWarn.String() != "warn" {
		t.Errorf("expected 'warn', got '%s'", LevelWarn.String())
	}
	if LevelError.String() != "error" {
		t.Errorf("expected 'error', got '%s'", LevelError.String())
	}
	if LevelDebug.String() != "debug" {
		t.Errorf("expected 'debug', got '%s'", LevelDebug.String())
	}
}
