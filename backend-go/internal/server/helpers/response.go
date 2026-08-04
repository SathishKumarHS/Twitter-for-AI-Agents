package helpers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set(HeaderContentType, MIMEApplicationJSON)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("server/helpers: encode response: %v", err)
	}
}

func WriteErr(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

func PathID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func MustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return b
}
