package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/agent-feed/backend-go/internal/server/helpers"
)

func (h *Handlers) Feed(w http.ResponseWriter, r *http.Request) {
	limit := helpers.DefaultFeedLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 && v <= helpers.MaxFeedLimit {
			limit = v
		}
	}

	if h.redisClient != nil {
		if val := h.redisClient.GetFeed(r.Context(), limit); val != nil {
			w.Header().Set(helpers.HeaderContentType, helpers.MIMEApplicationJSON)
			w.Header().Set(helpers.HeaderXCache, helpers.CacheHit)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(*val))
			return
		}
	}

	list, err := h.db.PostDB.Recent(r.Context(), limit)
	if err != nil {
		helpers.WriteErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	if h.redisClient != nil {
		if b, err := json.Marshal(list); err == nil {
			_ = h.redisClient.SetFeed(r.Context(), limit, b)
		}
	}

	w.Header().Set(helpers.HeaderXCache, helpers.CacheMiss)
	helpers.WriteJSON(w, http.StatusOK, list)
}

func (h *Handlers) Thread(w http.ResponseWriter, r *http.Request) {
	rootID, err := helpers.PathID(r)
	if err != nil {
		helpers.WriteErr(w, http.StatusBadRequest, "invalid thread id")
		return
	}
	list, err := h.db.PostDB.ThreadTree(r.Context(), rootID)
	if err != nil {
		helpers.WriteErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.WriteJSON(w, http.StatusOK, list)
}
