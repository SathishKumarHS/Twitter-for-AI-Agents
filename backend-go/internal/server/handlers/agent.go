package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	dberrors "github.com/twitter-for-ai-agents/backend-go/internal/db/db_errors"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
	"github.com/twitter-for-ai-agents/backend-go/internal/events"
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
	req.Handle = normalizeHandle(req.Handle)

	if msg := validateAgent(req); msg != "" {
		helpers.WriteErr(w, http.StatusBadRequest, msg)
		return
	}

	if existing, err := h.db.AgentsDB.GetByHandle(r.Context(), req.Handle); err == nil && existing != nil {
		helpers.WriteErr(w, http.StatusConflict, "handle already in use")
		return
	} else if err != nil && !errors.Is(err, dberrors.ErrNotFound) {
		helpers.WriteErr(w, http.StatusInternalServerError, "check handle: "+err.Error())
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
	slog.Info("Created new AI Agent", "agent_id", a.ID, "agent_name", a.Name, "agent_handle", a.Handle)
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
	slog.Info("Agent %s %s", "agent_id", agent.ID, "agent_handle", agent.Handle, "action", action)
	h.hub.Publish(events.Event{Type: events.EventAgentUpdated, Data: map[string]any{"id": id, "active": active}})
	helpers.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "active": active})
}

func (h *Handlers) PauseAgent(w http.ResponseWriter, r *http.Request) {
	h.setAgentActive(w, r, false)
}

func (h *Handlers) ResumeAgent(w http.ResponseWriter, r *http.Request) {
	h.setAgentActive(w, r, true)
}

const (
	maxHandleLen    = 30
	maxNameLen      = 120
	maxPersonaLen   = 2000
	maxListEntries  = 20
	maxListItemLen  = 500
	maxAvatarURLLen = 500
)

func normalizeHandle(handle string) string {
	return strings.TrimPrefix(strings.TrimSpace(handle), "@")
}

func validateAgent(req CreateAgentReq) string {
	if req.Handle == "" {
		return "handle is required"
	}
	if req.Name == "" {
		return "name is required"
	}
	if req.Persona == "" {
		return "persona is required"
	}
	if len(req.Handle) > maxHandleLen {
		return fmt.Sprintf("handle must be at most %d characters", maxHandleLen)
	}
	for _, ch := range req.Handle {
		if !(ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '_') {
			return "handle may only contain letters, digits, and underscores"
		}
	}
	if len(req.Name) > maxNameLen {
		return fmt.Sprintf("name must be at most %d characters", maxNameLen)
	}
	if len(req.Persona) > maxPersonaLen {
		return fmt.Sprintf("persona must be at most %d characters", maxPersonaLen)
	}
	if msg := validateStringList("goals", req.Goals); msg != "" {
		return msg
	}
	if msg := validateStringList("policies", req.Policies); msg != "" {
		return msg
	}
	if msg := validateStringList("interests", req.Interests); msg != "" {
		return msg
	}
	if req.AvatarURL != "" {
		if len(req.AvatarURL) > maxAvatarURLLen {
			return fmt.Sprintf("avatar_url must be at most %d characters", maxAvatarURLLen)
		}
		if u, err := url.Parse(req.AvatarURL); err != nil || (u.Scheme != "http" && u.Scheme != "https") {
			return "avatar_url must be a valid http(s) URL"
		}
	}
	return ""
}

func validateStringList(field string, items []string) string {
	if len(items) > maxListEntries {
		return fmt.Sprintf("%s must have at most %d entries", field, maxListEntries)
	}
	for _, it := range items {
		if len(it) > maxListItemLen {
			return fmt.Sprintf("%s entries must be at most %d characters", field, maxListItemLen)
		}
	}
	return ""
}
