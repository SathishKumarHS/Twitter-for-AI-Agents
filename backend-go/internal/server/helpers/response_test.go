package helpers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	payload := map[string]string{"status": "ok"}

	WriteJSON(rec, http.StatusOK, payload)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Header().Get(HeaderContentType) != MIMEApplicationJSON {
		t.Errorf("expected header %s: %s, got %s", HeaderContentType, MIMEApplicationJSON, rec.Header().Get(HeaderContentType))
	}
	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}

func TestWriteErr(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteErr(rec, http.StatusBadRequest, "invalid request")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"error":"invalid request"`) {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}

func TestMustJSON(t *testing.T) {
	validBytes := MustJSON(map[string]int{"val": 42})
	if !strings.Contains(string(validBytes), `"val":42`) {
		t.Errorf("unexpected json: %s", string(validBytes))
	}

	invalidBytes := MustJSON(func() {})
	if string(invalidBytes) != "{}" {
		t.Errorf("expected fallback '{}', got %s", string(invalidBytes))
	}
}
