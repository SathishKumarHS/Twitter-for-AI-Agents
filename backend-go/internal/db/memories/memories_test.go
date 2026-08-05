package memories

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
)

func setupTestDB(t *testing.T) MemoriesInterface {
	t.Helper()
	g, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}
	if err := g.AutoMigrate(&model.Memory{}); err != nil {
		t.Fatalf("failed to migrate memories model: %v", err)
	}
	return NewMemoriesDB(g)
}

func TestMemoriesDB_Operations(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	agentID := int64(10)

	memories0, err := db.List(ctx, agentID, 10)
	if err != nil || len(memories0) != 0 {
		t.Fatalf("expected 0 memories initially, got %d (err: %v)", len(memories0), err)
	}

	if err := db.Add(ctx, agentID, "", "Learned about Go concurrency"); err != nil {
		t.Fatalf("failed to add memory 1: %v", err)
	}

	if err := db.Add(ctx, agentID, "interaction", "Replied to @tech_bot"); err != nil {
		t.Fatalf("failed to add memory 2: %v", err)
	}

	list, err := db.List(ctx, agentID, 10)
	if err != nil || len(list) != 2 {
		t.Fatalf("expected 2 memories, got %d (err: %v)", len(list), err)
	}

	if list[0].Summary != "Replied to @tech_bot" || list[0].Kind != "interaction" {
		t.Errorf("unexpected latest memory: %+v", list[0])
	}

	if list[1].Summary != "Learned about Go concurrency" || list[1].Kind != "observation" {
		t.Errorf("unexpected first memory: %+v", list[1])
	}
}
