package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return &Server{logger: logger}
}

func TestHandleInit_BadUUID(t *testing.T) {
	s := newTestServer(t)
	body, _ := json.Marshal(map[string]any{"request_uuid": "not-a-uuid", "ttl_seconds": 60})
	req := httptest.NewRequest("POST", "/v1/init_request", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.initHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rec.Code)
	}
	var env ErrorEnvelope
	json.NewDecoder(rec.Body).Decode(&env)
	if env.ErrorType != string(ErrSchemaValidationFailed) {
		t.Errorf("error_type: %s", env.ErrorType)
	}
}

func TestHandleInit_BadTTL(t *testing.T) {
	s := newTestServer(t)
	body, _ := json.Marshal(map[string]any{"request_uuid": "f47ac10b-58cc-4372-a567-0e02b2c3d479", "ttl_seconds": 0})
	req := httptest.NewRequest("POST", "/v1/init_request", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	s.initHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rec.Code)
	}
}

func TestIsUUID(t *testing.T) {
	good := []string{"f47ac10b-58cc-4372-a567-0e02b2c3d479", "00000000-0000-0000-0000-000000000000"}
	bad := []string{"", "F47AC10B-58CC-4372-A567-0E02B2C3D479", "f47ac10b58cc4372a5670e02b2c3d479", "uuid"}
	for _, s := range good {
		if !isUUID(s) {
			t.Errorf("expected good UUID: %s", s)
		}
	}
	for _, s := range bad {
		if isUUID(s) {
			t.Errorf("expected bad UUID: %s", s)
		}
	}
}
