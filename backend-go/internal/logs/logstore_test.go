package logs

import (
	"fmt"
	"testing"

	"github.com/twitter-for-ai-agents/backend-go/internal/events"
)

func TestLogStoreRingBufferLimit(t *testing.T) {
	hub := events.NewHub()
	store := Init(hub)

	total := DefaultLogBufferCapacity + 10
	for i := 1; i <= total; i++ {
		store.add(LevelInfo, SystemNamespace, 0, "", "", fmt.Sprintf("Log message %d", i))
	}

	recent := store.Recent(0)
	if len(recent) != DefaultLogBufferCapacity {
		t.Fatalf("expected buffer length capped at %d, got %d", DefaultLogBufferCapacity, len(recent))
	}

	if recent[0].Message != "Log message 11" {
		t.Fatalf("expected oldest entry 'Log message 11', got '%s'", recent[0].Message)
	}

	if recent[len(recent)-1].Message != fmt.Sprintf("Log message %d", total) {
		t.Fatalf("expected latest entry 'Log message %d', got '%s'", total, recent[len(recent)-1].Message)
	}
}

func TestLogStoreRecentLimitHandling(t *testing.T) {
	hub := events.NewHub()
	store := Init(hub)

	for i := 1; i <= 5; i++ {
		store.add(LevelInfo, SystemNamespace, 0, "", "", fmt.Sprintf("Msg %d", i))
	}

	lastTwo := store.Recent(2)
	if len(lastTwo) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(lastTwo))
	}
	if lastTwo[0].Message != "Msg 4" || lastTwo[1].Message != "Msg 5" {
		t.Fatalf("unexpected recent entries: %+v", lastTwo)
	}

	overLimit := store.Recent(100)
	if len(overLimit) != 5 {
		t.Fatalf("expected 5 entries for overLimit, got %d", len(overLimit))
	}

	negativeLimit := store.Recent(-5)
	if len(negativeLimit) != 5 {
		t.Fatalf("expected 5 entries for negative limit, got %d", len(negativeLimit))
	}

	if GetDefault() != store {
		t.Fatalf("expected GetDefault() to return initialized store")
	}
}
