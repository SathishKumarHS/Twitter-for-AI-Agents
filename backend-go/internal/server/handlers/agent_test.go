package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/twitter-for-ai-agents/backend-go/internal/db"
	dberrors "github.com/twitter-for-ai-agents/backend-go/internal/db/db_errors"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/mocks"
)

func TestCreateAgentSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgents := mocks.NewMockAgentsInterface(ctrl)
	mockHub := mocks.NewMockHubInterface(ctrl)
	mockRedis := mocks.NewMockRedisInterface(ctrl)

	mockAgents.EXPECT().GetByHandle(gomock.Any(), "brand_bot").Return(nil, dberrors.ErrNotFound)
	mockAgents.EXPECT().Create(gomock.Any(), gomock.Any()).Return(int64(100), nil)
	mockAgents.EXPECT().Get(gomock.Any(), int64(100)).Return(&model.Agent{
		ID:        100,
		Handle:    "brand_bot",
		Name:      "Brand Bot",
		Persona:   "Friendly brand",
		Active:    true,
		Goals:     []string{"Engage"},
		Policies:  []string{"Polite"},
		Interests: []string{"Tech"},
	}, nil)

	mockHub.EXPECT().Publish(gomock.Any()).AnyTimes()

	h := New(&db.DB{AgentsDB: mockAgents}, mockHub, mockRedis)

	reqBody := CreateAgentReq{
		Handle:    "brand_bot",
		Name:      "Brand Bot",
		Persona:   "Friendly brand",
		Goals:     []string{"Engage"},
		Policies:  []string{"Polite"},
		Interests: []string{"Tech"},
	}
	b, _ := json.Marshal(reqBody)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/agents", bytes.NewReader(b))

	h.CreateAgent(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), `"handle":"brand_bot"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestCreateAgentBadRequest(t *testing.T) {
	h := New(&db.DB{}, nil, nil)

	reqBody := CreateAgentReq{Handle: "bot"}
	b, _ := json.Marshal(reqBody)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/agents", bytes.NewReader(b))

	h.CreateAgent(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request, got %d", rec.Code)
	}
}

func TestCreateAgentValidation(t *testing.T) {
	cases := []struct {
		name string
		req  CreateAgentReq
	}{
		{"empty handle", CreateAgentReq{Name: "Name", Persona: "Persona"}},
		{"empty name", CreateAgentReq{Handle: "bot", Persona: "Persona"}},
		{"empty persona", CreateAgentReq{Handle: "bot", Name: "Name"}},
		{"invalid handle chars", CreateAgentReq{Handle: "bad@handle!", Name: "Name", Persona: "Persona"}},
		{"handle too long", CreateAgentReq{Handle: strings.Repeat("a", 31), Name: "Name", Persona: "Persona"}},
		{"name too long", CreateAgentReq{Handle: "bot", Name: strings.Repeat("n", 121), Persona: "Persona"}},
		{"too many goals", CreateAgentReq{Handle: "bot", Name: "Name", Persona: "Persona", Goals: make([]string, 21)}},
		{"goal too long", CreateAgentReq{Handle: "bot", Name: "Name", Persona: "Persona", Goals: []string{strings.Repeat("g", 501)}}},
		{"invalid avatar", CreateAgentReq{Handle: "bot", Name: "Name", Persona: "Persona", AvatarURL: "not-a-url"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := New(&db.DB{}, nil, nil)
			b, _ := json.Marshal(tc.req)
			rec := httptest.NewRecorder()
			h.CreateAgent(rec, httptest.NewRequest("POST", "/api/agents", bytes.NewReader(b)))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCreateAgentHandleNormalized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgents := mocks.NewMockAgentsInterface(ctrl)
	mockHub := mocks.NewMockHubInterface(ctrl)

	mockAgents.EXPECT().GetByHandle(gomock.Any(), "brand_bot").Return(nil, dberrors.ErrNotFound)
	mockAgents.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ interface{}, a model.Agent) (int64, error) {
			if a.Handle != "brand_bot" {
				t.Fatalf("expected normalized handle 'brand_bot', got %q", a.Handle)
			}
			return int64(100), nil
		})
	mockAgents.EXPECT().Get(gomock.Any(), int64(100)).Return(&model.Agent{ID: 100, Handle: "brand_bot"}, nil)
	mockHub.EXPECT().Publish(gomock.Any()).AnyTimes()

	h := New(&db.DB{AgentsDB: mockAgents}, mockHub, nil)

	reqBody := CreateAgentReq{Handle: "  @brand_bot ", Name: "Brand Bot", Persona: "Friendly"}
	b, _ := json.Marshal(reqBody)

	rec := httptest.NewRecorder()
	h.CreateAgent(rec, httptest.NewRequest("POST", "/api/agents", bytes.NewReader(b)))

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateAgentDuplicateHandle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgents := mocks.NewMockAgentsInterface(ctrl)
	mockAgents.EXPECT().GetByHandle(gomock.Any(), "brand_bot").Return(&model.Agent{ID: 1, Handle: "brand_bot"}, nil)

	h := New(&db.DB{AgentsDB: mockAgents}, nil, nil)

	reqBody := CreateAgentReq{Handle: "brand_bot", Name: "Brand Bot", Persona: "Friendly"}
	b, _ := json.Marshal(reqBody)

	rec := httptest.NewRecorder()
	h.CreateAgent(rec, httptest.NewRequest("POST", "/api/agents", bytes.NewReader(b)))

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPauseAndResumeAgent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgents := mocks.NewMockAgentsInterface(ctrl)
	mockHub := mocks.NewMockHubInterface(ctrl)
	mockRedis := mocks.NewMockRedisInterface(ctrl)

	mockAgents.EXPECT().Get(gomock.Any(), int64(1)).Return(&model.Agent{ID: 1, Handle: "bot1", Active: true}, nil).Times(2)
	mockAgents.EXPECT().SetActive(gomock.Any(), int64(1), false).Return(nil)
	mockAgents.EXPECT().SetActive(gomock.Any(), int64(1), true).Return(nil)

	mockRedis.EXPECT().SetAgentActive(gomock.Any(), int64(1), false).Return(nil)
	mockRedis.EXPECT().SetAgentActive(gomock.Any(), int64(1), true).Return(nil)

	mockHub.EXPECT().Publish(gomock.Any()).Times(2)

	h := New(&db.DB{AgentsDB: mockAgents}, mockHub, mockRedis)

	// Test PauseAgent
	reqPause := httptest.NewRequest("POST", "/api/agents/1/pause", nil)
	reqPause.SetPathValue("id", "1")
	recPause := httptest.NewRecorder()

	h.PauseAgent(recPause, reqPause)

	if recPause.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK on pause, got %d", recPause.Code)
	}

	// Test ResumeAgent
	reqResume := httptest.NewRequest("POST", "/api/agents/1/resume", nil)
	reqResume.SetPathValue("id", "1")
	recResume := httptest.NewRecorder()

	h.ResumeAgent(recResume, reqResume)

	if recResume.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK on resume, got %d", recResume.Code)
	}
}
