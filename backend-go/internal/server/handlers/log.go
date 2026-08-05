package handlers

import (
	"net/http"
	"strconv"

	"github.com/twitter-for-ai-agents/backend-go/internal/logs"
	"github.com/twitter-for-ai-agents/backend-go/internal/server/helpers"
)

func (h *Handlers) GetLogs(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := helpers.DefaultLogsLimit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	logStore := logs.GetDefault()
	if logStore == nil {
		helpers.WriteJSON(w, http.StatusOK, []logs.LogEntry{})
		return
	}
	helpers.WriteJSON(w, http.StatusOK, logStore.Recent(limit))
}
