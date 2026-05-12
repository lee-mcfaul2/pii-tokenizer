package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type initRequestBody struct {
	RequestUUID string `json:"request_uuid"`
	TTLSeconds  int    `json:"ttl_seconds"`
}

type initResponseBody struct {
	RequestUUID string `json:"request_uuid"`
	ExpiresAt   string `json:"expires_at"`
}

var uuidLen = len("00000000-0000-0000-0000-000000000000")

func isUUID(s string) bool {
	if len(s) != uuidLen {
		return false
	}
	for i, c := range s {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				return false
			}
		}
	}
	return true
}

func (s *Server) initHandler(w http.ResponseWriter, r *http.Request) {
	var body initRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, r, ErrSchemaValidationFailed, "invalid JSON: "+err.Error())
		return
	}
	if !isUUID(body.RequestUUID) {
		WriteError(w, r, ErrSchemaValidationFailed, "request_uuid must be a lowercase RFC 4122 UUID")
		return
	}
	if body.TTLSeconds < 1 || body.TTLSeconds > 86400 {
		WriteError(w, r, ErrSchemaValidationFailed, "ttl_seconds must be in [1, 86400]")
		return
	}

	ttl := time.Duration(body.TTLSeconds) * time.Second
	res, err := s.scope.InitRequest(r.Context(), body.RequestUUID, ttl)
	if err != nil {
		if isRedisErr(err) {
			WriteError(w, r, ErrRedisUnavailable, err.Error())
			return
		}
		if isKMasterErr(err) {
			WriteError(w, r, ErrKMasterUnavailable, err.Error())
			return
		}
		WriteError(w, r, ErrInternal, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if res.Created {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_ = json.NewEncoder(w).Encode(initResponseBody{
		RequestUUID: body.RequestUUID,
		ExpiresAt:   res.ExpiresAt.UTC().Format(time.RFC3339),
	})
}

func isRedisErr(err error) bool { return strings.Contains(err.Error(), "redis") }
func isKMasterErr(err error) bool {
	return strings.Contains(err.Error(), "wrap K_request") ||
		strings.Contains(err.Error(), "unwrap K_request")
}
