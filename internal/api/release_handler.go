package api

import (
	"encoding/json"
	"net/http"
)

type releaseBody struct {
	RequestUUID string `json:"request_uuid"`
}

func (s *Server) releaseHandler(w http.ResponseWriter, r *http.Request) {
	var body releaseBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, r, ErrSchemaValidationFailed, err.Error())
		return
	}
	if !isUUID(body.RequestUUID) {
		WriteError(w, r, ErrSchemaValidationFailed, "request_uuid must be a UUID")
		return
	}
	if err := s.scope.ReleaseRequest(r.Context(), body.RequestUUID); err != nil {
		mapScopeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
