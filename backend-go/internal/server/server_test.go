package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/twitter-for-ai-agents/backend-go/internal/db"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/mocks"
	"go.uber.org/mock/gomock"
)

func TestHealthzEndpoint(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgents := mocks.NewMockAgentsInterface(ctrl)
	mockPosts := mocks.NewMockPostInterface(ctrl)
	mockHub := mocks.NewMockHubInterface(ctrl)

	database := &db.DB{
		AgentsDB: mockAgents,
		PostDB:   mockPosts,
	}

	srv := New(":8080", database, mockHub, nil)

	mux := http.NewServeMux()
	srv.routes(mux)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/healthz", nil)

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	if strings.TrimSpace(rec.Body.String()) != "ok" {
		t.Errorf("expected body 'ok', got '%s'", rec.Body.String())
	}
}

func TestListAgentsRoute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgents := mocks.NewMockAgentsInterface(ctrl)
	mockPosts := mocks.NewMockPostInterface(ctrl)
	mockHub := mocks.NewMockHubInterface(ctrl)

	mockAgents.EXPECT().ListAll(gomock.Any()).Return([]model.Agent{
		{ID: 1, Handle: "bot1", Name: "Bot One"},
		{ID: 2, Handle: "bot2", Name: "Bot Two"},
	}, nil)

	database := &db.DB{
		AgentsDB: mockAgents,
		PostDB:   mockPosts,
	}

	srv := New(":8080", database, mockHub, nil)

	mux := http.NewServeMux()
	srv.routes(mux)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/agents", nil)

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"handle":"bot1"`) || !strings.Contains(rec.Body.String(), `"handle":"bot2"`) {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}
