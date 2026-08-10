package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/twitter-for-ai-agents/backend-go/internal/db"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/internal/server/helpers"
	"github.com/twitter-for-ai-agents/backend-go/mocks"
)

func TestFeedCacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cachedJSON := `[{"id": 1, "body": "cached post"}]`
	mockRedis := mocks.NewMockRedisInterface(ctrl)

	mockRedis.EXPECT().GetFeed(gomock.Any(), 50).Return(&cachedJSON)

	h := New(&db.DB{}, nil, mockRedis, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/posts?limit=50", nil)

	h.Feed(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	if rec.Header().Get(helpers.HeaderXCache) != helpers.CacheHit {
		t.Fatalf("expected X-Cache header 'HIT', got '%s'", rec.Header().Get(helpers.HeaderXCache))
	}

	if rec.Body.String() != cachedJSON {
		t.Fatalf("expected body '%s', got '%s'", cachedJSON, rec.Body.String())
	}
}

func TestFeedCacheMiss(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPosts := mocks.NewMockPostInterface(ctrl)
	mockRedis := mocks.NewMockRedisInterface(ctrl)

	mockRedis.EXPECT().GetFeed(gomock.Any(), 50).Return(nil)
	mockPosts.EXPECT().Recent(gomock.Any(), 50).Return([]model.Post{
		{ID: 10, AgentID: 1, Body: "fresh post"},
	}, nil)
	mockRedis.EXPECT().SetFeed(gomock.Any(), 50, gomock.Any()).Return(nil)

	h := New(&db.DB{PostDB: mockPosts}, nil, mockRedis, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/posts?limit=50", nil)

	h.Feed(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	if rec.Header().Get(helpers.HeaderXCache) != helpers.CacheMiss {
		t.Fatalf("expected X-Cache header 'MISS', got '%s'", rec.Header().Get(helpers.HeaderXCache))
	}

	if !strings.Contains(rec.Body.String(), `"body":"fresh post"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestThreadRoute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPosts := mocks.NewMockPostInterface(ctrl)

	mockPosts.EXPECT().ThreadTree(gomock.Any(), int64(1)).Return([]model.Post{
		{ID: 1, AgentID: 1, Body: "root post"},
		{ID: 2, AgentID: 2, Body: "reply post"},
	}, nil)

	h := New(&db.DB{PostDB: mockPosts}, nil, nil, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/posts/thread/1", nil)
	req.SetPathValue("id", "1")

	h.Thread(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"body":"root post"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}
