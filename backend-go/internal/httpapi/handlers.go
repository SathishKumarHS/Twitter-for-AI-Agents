package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/agent-feed/backend-go/internal/db/agents"
	dberrors "github.com/agent-feed/backend-go/internal/db/db_errors"
	"github.com/agent-feed/backend-go/internal/events"
	"github.com/agent-feed/backend-go/internal/logs"
)

// writeJSON marshals v and writes it with the given status.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("httpapi: encode response: %v", err)
	}
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func pathID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func (s *Server) handleListAgents(w http.ResponseWriter, r *http.Request) {
	list, err := s.db.AgentsDB.ListAll(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "list agents: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

type createAgentReq struct {
	Handle    string   `json:"handle"`
	Name      string   `json:"name"`
	Persona   string   `json:"persona"`
	Goals     []string `json:"goals"`
	Policies  []string `json:"policies"`
	Interests []string `json:"interests"`
	AvatarURL string   `json:"avatar_url"`
}

func (s *Server) handleCreateAgent(w http.ResponseWriter, r *http.Request) {
	var req createAgentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	if req.Handle == "" || req.Name == "" || req.Persona == "" {
		writeErr(w, http.StatusBadRequest, "handle, name, persona are required")
		return
	}
	id, err := s.db.AgentsDB.Create(r.Context(), agents.Agent{
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
		writeErr(w, http.StatusInternalServerError, "create agent: "+err.Error())
		return
	}
	a, err := s.db.AgentsDB.Get(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "fetch created agent: "+err.Error())
		return
	}
	s.publish(events.Event{Type: "agent:new", Data: a})
	logs.AgentInfo(a.ID, a.Name, a.Handle, "system", fmt.Sprintf("Created new AI Agent @%s (%s)", a.Handle, a.Name))
	writeJSON(w, http.StatusCreated, a)
}

func (s *Server) handleGetAgent(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid agent id")
		return
	}
	a, err := s.db.AgentsDB.Get(r.Context(), id)
	if errors.Is(err, dberrors.ErrNotFound) {
		writeErr(w, http.StatusNotFound, "agent not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (s *Server) handleAgentPosts(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid agent id")
		return
	}
	list, err := s.db.PostDB.Timeline(r.Context(), id, 50)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handlePauseAgent(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid agent id")
		return
	}
	agent, err := s.db.AgentsDB.Get(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "agent not found")
		return
	}
	if err := s.db.AgentsDB.SetActive(r.Context(), id, false); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if s.cache != nil {
		_ = s.cache.SetAgentActive(r.Context(), id, false)
	}
	logs.Info("system", fmt.Sprintf("Agent #%d (@%s) paused", agent.ID, agent.Handle))
	s.hub.Publish(events.Event{Type: "agent:updated", Data: map[string]any{"id": id, "active": false}})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "active": false})
}

func (s *Server) handleResumeAgent(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid agent id")
		return
	}
	agent, err := s.db.AgentsDB.Get(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "agent not found")
		return
	}
	if err := s.db.AgentsDB.SetActive(r.Context(), id, true); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if s.cache != nil {
		_ = s.cache.SetAgentActive(r.Context(), id, true)
	}
	logs.Info("system", fmt.Sprintf("Agent #%d (@%s) resumed", agent.ID, agent.Handle))
	s.hub.Publish(events.Event{Type: "agent:updated", Data: map[string]any{"id": id, "active": true}})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "active": true})
}

func (s *Server) handleFeed(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}

	if s.cache != nil {
		if val := s.cache.GetFeed(r.Context(), limit); val != nil {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(*val))
			return
		}
	}

	list, err := s.db.PostDB.Recent(r.Context(), limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	if s.cache != nil {
		if b, err := json.Marshal(list); err == nil {
			_ = s.cache.SetFeed(r.Context(), limit, b)
		}
	}

	w.Header().Set("X-Cache", "MISS")
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleThread(w http.ResponseWriter, r *http.Request) {
	rootID, err := pathID(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid thread id")
		return
	}
	list, err := s.db.PostDB.ThreadTree(r.Context(), rootID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	logStore := logs.GetDefault()
	if logStore == nil {
		writeJSON(w, http.StatusOK, []logs.LogEntry{})
		return
	}
	writeJSON(w, http.StatusOK, logStore.Recent(limit))
}

func (s *Server) publish(ev events.Event) {
	if s.hub != nil {
		s.hub.Publish(ev)
	}
}
