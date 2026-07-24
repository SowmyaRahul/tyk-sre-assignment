package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SowmyaRahul/tyk-sre-assignment/internal/isolation"
)

// handleIsolationCollection serves POST (create) and GET (list).
func (s *Server) handleIsolationCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.createIsolation(w, r)
	case http.MethodGet:
		s.listIsolation(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST or GET")
	}
}

// handleIsolationItem serves DELETE /isolation/{id}.
func (s *Server) handleIsolationItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use DELETE")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/isolation/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "missing isolation id")
		return
	}
	n, err := s.mgr.Delete(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "delete_failed", err.Error())
		return
	}
	if n == 0 {
		writeError(w, http.StatusNotFound, "not_found", "no isolation with that id")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"id": id, "deleted": n})
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