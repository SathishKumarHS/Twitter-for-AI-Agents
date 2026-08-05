package posts

import (
	"context"
	"reflect"
	"testing"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/mocks"
	"go.uber.org/mock/gomock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPostsDB(t *testing.T, lastSeenMock *mocks.MockLastSeenInterface) PostInterface {
	t.Helper()
	g, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}
	if err := g.AutoMigrate(&model.Agent{}, &model.Post{}); err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}
	return NewPostsDB(g, lastSeenMock)
}

func TestExtractMentionHandles(t *testing.T) {
	body := "Hello @alice and @bob! Check out @charlie."
	handles := extractMentionHandles(body)
	expected := []string{"alice", "bob", "charlie"}

	if !reflect.DeepEqual(handles, expected) {
		t.Errorf("expected handles %v, got %v", expected, handles)
	}
}

func TestPostsDB_CreateAndRetrieve(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLastSeen := mocks.NewMockLastSeenInterface(ctrl)
	db := setupPostsDB(t, mockLastSeen)
	ctx := context.Background()

	post1, err := db.Create(ctx, 1, "First main post", 0)
	if err != nil || post1.ID == 0 {
		t.Fatalf("failed to create main post: %v", err)
	}
	if post1.Kind != "post" || post1.ParentID != nil {
		t.Fatalf("expected main post kind 'post' and nil ParentID")
	}

	post2, err := db.Create(ctx, 2, "Reply to first post", post1.ID)
	if err != nil || post2.ID == 0 {
		t.Fatalf("failed to create reply post: %v", err)
	}
	if post2.Kind != "reply" || post2.ParentID == nil || *post2.ParentID != post1.ID {
		t.Fatalf("expected reply post ParentID = %d", post1.ID)
	}
	if post2.RootID == nil || *post2.RootID != post1.ID {
		t.Fatalf("expected reply post RootID = %d", post1.ID)
	}

	recent, err := db.Recent(ctx, 10)
	if err != nil || len(recent) != 2 {
		t.Fatalf("expected 2 recent posts, got %d (err: %v)", len(recent), err)
	}

	timeline, err := db.Timeline(ctx, 1, 10)
	if err != nil || len(timeline) != 1 || timeline[0].ID != post1.ID {
		t.Fatalf("expected 1 post in agent 1 timeline, got %d", len(timeline))
	}

	threadTree, err := db.ThreadTree(ctx, post1.ID)
	if err != nil || len(threadTree) != 1 {
		t.Fatalf("expected 1 root post in ThreadTree, got %d (err: %v)", len(threadTree), err)
	}
	if len(threadTree[0].Comments) != 1 || threadTree[0].Comments[0].ID != post2.ID {
		t.Fatalf("expected 1 comment attached in thread tree")
	}

	lastMainTime, err := db.LastMainPost(ctx, 1)
	if err != nil || lastMainTime.IsZero() {
		t.Fatalf("expected non-zero LastMainPost time, got %v (err: %v)", lastMainTime, err)
	}

	mockLastSeen.EXPECT().LastSeen(gomock.Any(), int64(2)).Return(post1.ID)
	newPosts, err := db.NewPostsAfter(ctx, 2)
	if err != nil {
		t.Fatalf("failed to fetch new posts after: %v", err)
	}
	if len(newPosts) != 0 {
		t.Fatalf("expected 0 new posts for agent 2, got %d", len(newPosts))
	}
}
