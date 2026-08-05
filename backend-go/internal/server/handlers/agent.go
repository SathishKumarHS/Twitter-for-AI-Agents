package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	dberrors "github.com/twitter-for-ai-agents/backend-go/internal/db/db_errors"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/internal/events"
	"github.com/twitter-for-ai-agents/backend-go/internal/logs"
	"github.com/twitter-for-ai-agents/backend-go/internal/server/helpers"
)

type CreateAgentReq struct {
	Handle    string   `json:"handle"`
	Name      string   `json:"name"`
	Persona   string   `json:"persona"`
	Goals     []string `json:"goals"`
	Policies  []string `json:"policies"`
	Interests []string `json:"interests"`
	AvatarURL string   `json:"avatar_url"`
}

func (h *Handlers) ListAgents(w http.ResponseWriter, r *http.Request) {
	list, err := h.db.AgentsDB.ListAll(r.Context())
	if err != nil {
		helpers.WriteErr(w, http.StatusInternalServerError, "list agents: "+err.Error())
		return
	}
	helpers.WriteJSON(w, http.StatusOK, list)
}

func (h *Handlers) CreateAgent(w http.ResponseWriter, r *http.Request) {
	var req CreateAgentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteErr(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	if req.Handle == "" || req.Name == "" || req.Persona == "" {
		helpers.WriteErr(w, http.StatusBadRequest, "handle, name, persona are required")
		return
	}
	id, err := h.db.AgentsDB.Create(r.Context(), model.Agent{
		Handle:    req.Handle,
		Name:      req.Name,
		Persona:   req.Persona,
		Goals:     req.Goals,
		Policies:  req.Policies,
		Interests: req.Interests,
		AvatarURL: req.AvatarURL,
		Active:    true,
	})
	if err != nil {
		helpers.WriteErr(w, http.StatusInternalServerError, "create agent: "+err.Error())
		return
	}
	a, err := h.db.AgentsDB.Get(r.Context(), id)
	if err != nil {
		helpers.WriteErr(w, http.StatusInternalServerError, "fetch created agent: "+err.Error())
		return
	}
	h.publish(events.Event{Type: events.EventAgentNew, Data: a})
	logs.AgentInfo(a.ID, a.Name, a.Handle, logs.SystemNamespace, fmt.Sprintf("Created new AI Agent @%s (%s)", a.Handle, a.Name))
	helpers.WriteJSON(w, http.StatusCreated, a)
}

func (h *Handlers) GetAgent(w http.ResponseWriter, r *http.Request) {
	id, err := helpers.PathID(r)
	if err != nil {
		helpers.WriteErr(w, http.StatusBadRequest, "invalid agent id")
		return
	}
	a, err := h.db.AgentsDB.Get(r.Context(), id)
	if errors.Is(err, dberrors.ErrNotFound) {
		helpers.WriteErr(w, http.StatusNotFound, "agent not found")
		return
	}
	if err != nil {
		helpers.WriteErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.WriteJSON(w, http.StatusOK, a)
}

func (h *Handlers) AgentPosts(w http.ResponseWriter, r *http.Request) {
	id, err := helpers.PathID(r)
	if err != nil {
		helpers.WriteErr(w, http.StatusBadRequest, "invalid agent id")
		return
	}
	list, err := h.db.PostDB.Timeline(r.Context(), id, 50)
	if err != nil {
		helpers.WriteErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.WriteJSON(w, http.StatusOK, list)
}

// setAgentActive is a helper for pausing and resuming agents.
func (h *Handlers) setAgentActive(w http.ResponseWriter, r *http.Request, active bool) {
	id, err := helpers.PathID(r)
	if err != nil {
		helpers.WriteErr(w, http.StatusBadRequest, "invalid agent id")
		return
	}
	agent, err := h.db.AgentsDB.Get(r.Context(), id)
	if err != nil {
		helpers.WriteErr(w, http.StatusNotFound, "agent not found")
		return
	}
	if err := h.db.AgentsDB.SetActive(r.Context(), id, active); err != nil {
		helpers.WriteErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if h.redisClient != nil {
		_ = h.redisClient.SetAgentActive(r.Context(), id, active)
	}
	action := "paused"
	if active {
		action = "resumed"
	}
	logs.Info(logs.SystemNamespace, fmt.Sprintf("Agent #%d (@%s) %s", agent.ID, agent.Handle, action))
	h.hub.Publish(events.Event{Type: events.EventAgentUpdated, Data: map[string]any{"id": id, "active": active}})
	helpers.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "active": active})
}

func (h *Handlers) PauseAgent(w http.ResponseWriter, r *http.Request) {
	h.setAgentActive(w, r, false)
}

func (h *Handlers) ResumeAgent(w http.ResponseWriter, r *http.Request) {
	h.setAgentActive(w, r, true)
}
