package server

import (
	"net/http"

	"github.com/SowmyaRahul/tyk-sre-assignment/internal/health"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Server) handleDeployments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use GET")
		return
	}
	ns := r.URL.Query().Get("namespace") // "" = all namespaces
	list, err := s.cs.AppsV1().Deployments(ns).List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, health.Compute(list))
}
