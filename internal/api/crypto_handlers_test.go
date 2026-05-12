package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleTokenize_BadType(t *testing.T) {
	s := newTestServer(t)
	body, _ := json.Marshal(map[string]any{
		"request_uuid": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
		"type":         "BOGUS_TYPE",
		"plaintext":    "x",
	})
	req := httptest.NewRequest("POST", "/v1/tokenize", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.tokenizeHandler(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rec.Code)
	}
	var env ErrorEnvelope
	json.NewDecoder(rec.Body).Decode(&env)
	if env.ErrorType != string(ErrInvalidPIIType) {
		t.Errorf("error_type: %s", env.ErrorType)
	}
}

func TestHandleTokenize_BadUUID(t *testing.T) {
	s := newTestServer(t)
	body, _ := json.Marshal(map[string]any{
		"request_uuid": "not-a-uuid",
		"type":         "EMAIL",
		"plaintext":    "x",
	})
	req := httptest.NewRequest("POST", "/v1/tokenize", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.tokenizeHandler(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rec.Code)
	}
}

func TestHandleDetokenize_MissingToken(t *testing.T) {
	s := newTestServer(t)
	body, _ := json.Marshal(map[string]any{
		"request_uuid": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
		"token":        "",
	})
	req := httptest.NewRequest("POST", "/v1/detokenize", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.detokenizeHandler(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rec.Code)
	}
}
