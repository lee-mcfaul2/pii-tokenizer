package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/lee-mcfaul2/pii-tokenizer/internal/store"
)

type tokenizeBody struct {
	RequestUUID string `json:"request_uuid"`
	Type        string `json:"type"`
	Plaintext   string `json:"plaintext"`
}
type tokenizeResp struct {
	Token string `json:"token"`
}
type detokenizeBody struct {
	RequestUUID string `json:"request_uuid"`
	Token       string `json:"token"`
}
type detokenizeResp struct {
	Plaintext string `json:"plaintext"`
	Type      string `json:"type"`
}

var allowedPIITypes = map[string]struct{}{
	"EMAIL": {}, "PHONE": {}, "SSN": {}, "ADDRESS": {}, "POSTAL_CODE": {},
	"NAME": {}, "CREDIT_CARD": {}, "IBAN": {}, "IP_ADDRESS": {}, "DOB": {},
}

func (s *Server) tokenizeHandler(w http.ResponseWriter, r *http.Request) {
	var body tokenizeBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, r, ErrSchemaValidationFailed, err.Error())
		return
	}
	if !isUUID(body.RequestUUID) {
		WriteError(w, r, ErrSchemaValidationFailed, "request_uuid must be a UUID")
		return
	}
	if _, ok := allowedPIITypes[body.Type]; !ok {
		WriteError(w, r, ErrInvalidPIIType, "type not in enum: "+body.Type)
		return
	}

	tok, err := s.scope.Tokenize(r.Context(), body.RequestUUID, body.Type, []byte(body.Plaintext))
	if err != nil {
		mapScopeError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(tokenizeResp{Token: tok})
}

func (s *Server) detokenizeHandler(w http.ResponseWriter, r *http.Request) {
	var body detokenizeBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, r, ErrSchemaValidationFailed, err.Error())
		return
	}
	if !isUUID(body.RequestUUID) {
		WriteError(w, r, ErrSchemaValidationFailed, "request_uuid must be a UUID")
		return
	}
	if body.Token == "" {
		WriteError(w, r, ErrSchemaValidationFailed, "token required")
		return
	}

	piiType, pt, err := s.scope.Detokenize(r.Context(), body.RequestUUID, body.Token)
	if err != nil {
		if strings.Contains(err.Error(), "base32") ||
			strings.Contains(err.Error(), "token") ||
			strings.Contains(err.Error(), "aes_siv: invalid ciphertext") {
			WriteError(w, r, ErrAADMismatch, err.Error())
			return
		}
		mapScopeError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(detokenizeResp{Plaintext: string(pt), Type: piiType})
}

func mapScopeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, store.ErrScopeNotFound):
		WriteError(w, r, ErrScopeNotFound, err.Error())
	case errors.Is(err, store.ErrScopeExpired):
		WriteError(w, r, ErrScopeExpired, err.Error())
	case strings.Contains(err.Error(), "unwrap K_request"):
		WriteError(w, r, ErrKMasterUnavailable, err.Error())
	case strings.Contains(err.Error(), "redis"):
		WriteError(w, r, ErrRedisUnavailable, err.Error())
	case strings.Contains(err.Error(), "AAD"):
		WriteError(w, r, ErrAADMismatch, err.Error())
	default:
		WriteError(w, r, ErrInternal, err.Error())
	}
}
