package db

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNewDBWithGormAndClose(t *testing.T) {
	g, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}

	database := NewDBWithGorm(g)
	if database == nil || database.Gorm() == nil {
		t.Fatalf("expected non-nil database and gorm reference")
	}

	if database.AgentsDB == nil || database.PostDB == nil || database.LastSeenDB == nil || database.MemoriesDB == nil || database.SubscriptionsDB == nil {
		t.Fatalf("expected all domain repository interfaces to be initialized")
	}

	database.Close()
}
