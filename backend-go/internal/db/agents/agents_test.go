package agents

import (
	"context"
	"errors"
	"testing"

	dberrors "github.com/twitter-for-ai-agents/backend-go/internal/db/db_errors"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) AgentsInterface {
	t.Helper()
	g, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}
	if err := g.AutoMigrate(&model.Agent{}); err != nil {
		t.Fatalf("failed to migrate agent model: %v", err)
	}
	return NewAgentsDB(g)
}

func TestAgentsDB_Operations(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	id1, err := db.Create(ctx, model.Agent{
		Handle:  "bot_one",
		Name:    "Bot One",
		Persona: "Active bot",
		Active:  true,
	})
	if err != nil || id1 == 0 {
		t.Fatalf("failed to create agent 1: %v", err)
	}

	id2, err := db.Create(ctx, model.Agent{
		Handle:  "bot_two",
		Name:    "Bot Two",
		Persona: "Inactive bot",
		Active:  false,
	})
	if err != nil || id2 == 0 {
		t.Fatalf("failed to create agent 2: %v", err)
	}

	all, err := db.ListAll(ctx)
	if err != nil || len(all) != 2 {
		t.Fatalf("expected 2 agents in ListAll, got %d (err: %v)", len(all), err)
	}

	active, err := db.ListActive(ctx)
	if err != nil || len(active) != 1 || active[0].ID != id1 {
		t.Fatalf("expected 1 active agent (id %d), got %+v (err: %v)", id1, active, err)
	}

	a1, err := db.Get(ctx, id1)
	if err != nil || a1.Handle != "bot_one" {
		t.Fatalf("expected agent handle 'bot_one', got %+v (err: %v)", a1, err)
	}

	a1Handle, err := db.GetByHandle(ctx, "BOT_ONE")
	if err != nil || a1Handle.ID != id1 {
		t.Fatalf("expected agent ID %d for handle lookup, got %+v (err: %v)", id1, a1Handle, err)
	}

	_, errNotFound := db.Get(ctx, 9999)
	if !errors.Is(errNotFound, dberrors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for missing agent ID, got %v", errNotFound)
	}

	if err := db.SetActive(ctx, id1, false); err != nil {
		t.Fatalf("failed to deactivate agent 1: %v", err)
	}

	activeAfterPause, err := db.ListActive(ctx)
	if err != nil || len(activeAfterPause) != 0 {
		t.Fatalf("expected 0 active agents after pause, got %d", len(activeAfterPause))
	}
}
