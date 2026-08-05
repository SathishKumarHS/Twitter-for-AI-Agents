package subscriptions

import (
	"context"
	"reflect"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
)

func setupTestDB(t *testing.T) SubscriptionInterface {
	t.Helper()
	g, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}
	if err := g.AutoMigrate(&model.ThreadSubscriber{}); err != nil {
		t.Fatalf("failed to migrate thread_subscribers: %v", err)
	}
	return NewSubscriptionDB(g)
}

func TestSubscriptionDB_Operations(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	rootPostID := int64(100)

	subs0, err := db.ListSubscribers(ctx, rootPostID)
	if err != nil || len(subs0) != 0 {
		t.Fatalf("expected 0 subscribers, got %d (err: %v)", len(subs0), err)
	}

	if err := db.Subscribe(ctx, 10, rootPostID); err != nil {
		t.Fatalf("failed to subscribe agent 10: %v", err)
	}
	if err := db.Subscribe(ctx, 20, rootPostID); err != nil {
		t.Fatalf("failed to subscribe agent 20: %v", err)
	}

	if err := db.Subscribe(ctx, 10, rootPostID); err != nil {
		t.Fatalf("failed to idempotently subscribe agent 10: %v", err)
	}

	subs, err := db.ListSubscribers(ctx, rootPostID)
	if err != nil {
		t.Fatalf("failed to list subscribers: %v", err)
	}
	expected := []int64{10, 20}
	if !reflect.DeepEqual(subs, expected) {
		t.Fatalf("expected subscribers %v, got %v", expected, subs)
	}
}
