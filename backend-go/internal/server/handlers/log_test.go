package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/twitter-for-ai-agents/backend-go/internal/events"
	"github.com/twitter-for-ai-agents/backend-go/internal/logs"
)

func TestGetLogsHandler(t *testing.T) {
	hub := events.NewHub()
	_ = logs.Init(hub)

	logs.Info(logs.SystemNamespace, "System test message")

	h := New(nil, hub, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/logs?limit=10", nil)

	h.GetLogs(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "System test message") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}
