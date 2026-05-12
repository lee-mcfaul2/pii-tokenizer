package api

import (
	"encoding/json"
	"net/http"
)

type ErrorType string

const (
	ErrScopeNotFound          ErrorType = "SCOPE_NOT_FOUND"
	ErrScopeExpired           ErrorType = "SCOPE_EXPIRED"
	ErrAADMismatch            ErrorType = "AAD_MISMATCH"
	ErrSchemaValidationFailed ErrorType = "SCHEMA_VALIDATION_FAILED"
	ErrInvalidPIIType         ErrorType = "INVALID_PII_TYPE"
	ErrKMasterUnavailable     ErrorType = "KMASTER_UNAVAILABLE"
	ErrRedisUnavailable       ErrorType = "REDIS_UNAVAILABLE"
	ErrInternal               ErrorType = "INTERNAL_ERROR"
)

type ErrorEnvelope struct {
	ErrorType string `json:"error_type"`
	Retriable bool   `json:"retriable"`
	Message   string `json:"message"`
	TraceID   string `json:"trace_id,omitempty"`
}

func (e ErrorType) HTTPStatus() int {
	switch e {
	case ErrScopeNotFound:
		return http.StatusNotFound
	case ErrScopeExpired:
		return http.StatusGone
	case ErrAADMismatch, ErrSchemaValidationFailed, ErrInvalidPIIType:
		return http.StatusBadRequest
	case ErrKMasterUnavailable, ErrRedisUnavailable:
		return http.StatusServiceUnavailable
	case ErrInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func (e ErrorType) Retriable() bool {
	switch e {
	case ErrKMasterUnavailable, ErrRedisUnavailable, ErrInternal:
		return true
	default:
		return false
	}
}

func WriteError(w http.ResponseWriter, r *http.Request, et ErrorType, msg string) {
	traceID := r.Header.Get("X-Request-ID")
	env := ErrorEnvelope{
		ErrorType: string(et),
		Retriable: et.Retriable(),
		Message:   msg,
		TraceID:   traceID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(et.HTTPStatus())
	_ = json.NewEncoder(w).Encode(env)
}
