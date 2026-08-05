package lastseen

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
)

func setupTestDB(t *testing.T) LastSeenInterface {
	t.Helper()
	g, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}
	if err := g.AutoMigrate(&model.AgentSeen{}); err != nil {
		t.Fatalf("failed to migrate agent_seens: %v", err)
	}
	return NewLastSeenDB(g)
}

func TestLastSeenDB_Operations(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	agentID := int64(42)

	if seen := db.LastSeen(ctx, agentID); seen != 0 {
		t.Fatalf("expected initial lastSeen 0, got %d", seen)
	}

	if err := db.SetSeen(ctx, agentID, 100); err != nil {
		t.Fatalf("failed to set seen: %v", err)
	}

	if seen := db.LastSeen(ctx, agentID); seen != 100 {
		t.Fatalf("expected lastSeen 100, got %d", seen)
	}

	if err := db.SetSeen(ctx, agentID, 250); err != nil {
		t.Fatalf("failed to update seen: %v", err)
	}

	if seen := db.LastSeen(ctx, agentID); seen != 250 {
		t.Fatalf("expected updated lastSeen 250, got %d", seen)
	}
}
