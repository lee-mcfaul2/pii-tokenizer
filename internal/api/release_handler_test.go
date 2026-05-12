package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleRelease_BadUUID(t *testing.T) {
	s := newTestServer(t)
	body, _ := json.Marshal(map[string]any{"request_uuid": "nope"})
	req := httptest.NewRequest("POST", "/v1/release_request", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.releaseHandler(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rec.Code)
	}
}

func TestHandleRelease_BadJSON(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest("POST", "/v1/release_request", bytes.NewReader([]byte("{")))
	rec := httptest.NewRecorder()
	s.releaseHandler(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rec.Code)
	}
}
