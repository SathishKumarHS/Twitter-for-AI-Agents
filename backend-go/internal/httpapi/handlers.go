package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"fmt"
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
	list, err := s.db.AgentsDB.ListActive(r.Context())
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

func (s *Server) handleAgentMemories(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid agent id")
		return
	}
	list, err := s.db.MemoriesDB.List(r.Context(), id, 30)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleFeed(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}
	list, err := s.db.PostDB.Recent(r.Context(), limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

type followReq struct {
	FollowerID  int64 `json:"follower_id"`
	FollowingID int64 `json:"following_id"`
}

func (s *Server) handleCreateFollow(w http.ResponseWriter, r *http.Request) {
	var req followReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	if req.FollowerID == 0 || req.FollowingID == 0 {
		writeErr(w, http.StatusBadRequest, "follower_id and following_id are required")
		return
	}
	if req.FollowerID == req.FollowingID {
		writeErr(w, http.StatusBadRequest, "an agent cannot follow itself")
		return
	}
	if err := s.db.FollowsDB.Add(r.Context(), req.FollowerID, req.FollowingID); err != nil {
		writeErr(w, http.StatusInternalServerError, "create follow: "+err.Error())
		return
	}
	logs.Info("system", fmt.Sprintf("Agent #%d followed Agent #%d", req.FollowerID, req.FollowingID))
	writeJSON(w, http.StatusCreated, map[string]bool{"ok": true})
}

func (s *Server) handleDeleteFollow(w http.ResponseWriter, r *http.Request) {
	var req followReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	if req.FollowerID == 0 || req.FollowingID == 0 {
		writeErr(w, http.StatusBadRequest, "follower_id and following_id are required")
		return
	}
	if err := s.db.FollowsDB.Remove(r.Context(), req.FollowerID, req.FollowingID); err != nil {
		writeErr(w, http.StatusInternalServerError, "delete follow: "+err.Error())
		return
	}
	logs.Info("system", fmt.Sprintf("Agent #%d unfollowed Agent #%d", req.FollowerID, req.FollowingID))
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleAgentFollowers(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid agent id")
		return
	}
	list, err := s.db.FollowsDB.Followers(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleAgentFollowing(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid agent id")
		return
	}
	list, err := s.db.FollowsDB.Following(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
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
