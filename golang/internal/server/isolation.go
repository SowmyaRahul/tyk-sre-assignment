package server

import (
	"encoding/json"
	"net/http"

	"github.com/SowmyaRahul/tyk-sre-assignment/internal/isolation"
)

// handleIsolationCollection serves POST (create, gated) and GET (list, open).
func (s *Server) handleIsolationCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		if !s.authorized(r) {
			writeError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token required")
			return
		}
		s.createIsolation(w, r)
	case http.MethodGet:
		s.listIsolation(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST or GET")
	}
}

func (s *Server) createIsolation(w http.ResponseWriter, r *http.Request) {
	var req isolation.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if req.A.Namespace == "" || req.B.Namespace == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "both a.namespace and b.namespace are required")
		return
	}
	id, converged, err := s.mgr.Apply(r.Context(), req, "api")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "apply_failed", err.Error())
		return
	}
	status := http.StatusCreated
	if converged {
		status = http.StatusOK
	}
	writeJSON(w, status, map[string]interface{}{"id": id, "converged": converged})
}

func (s *Server) listIsolation(w http.ResponseWriter, r *http.Request) {
	items, err := s.mgr.List(r.Context())
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"isolations": items})
}
