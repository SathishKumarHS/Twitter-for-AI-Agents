package schema

import (
	"context"
	"testing"

	"github.com/twitter-for-ai-agents/backend-go/internal/db"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMigrate(t *testing.T) {
	g, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}

	database := db.NewDBWithGorm(g)
	ctx := context.Background()

	if err := Migrate(ctx, database); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}
}
