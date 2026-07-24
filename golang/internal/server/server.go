package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SowmyaRahul/tyk-sre-assignment/internal/k8s"
	"k8s.io/client-go/kubernetes"
)

// Server wires HTTP handlers to the k8s-facing components.
type Server struct {
	cs     kubernetes.Interface
	pinger k8s.Pinger
}

// New constructs a Server.
func New(cs kubernetes.Interface, pinger k8s.Pinger) *Server {
	return &Server{cs: cs, pinger: pinger}
}

// Handler returns the configured router.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/readyz", s.handleReadyz)
	mux.HandleFunc("/deployments", s.handleDeployments) // commit_6
	return mux
}

// handleHealthz is the liveness probe. 
func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		fmt.Println("failed writing to response")
	}
}

// handleReadyz is the readiness probe: it reports API-server connectivity (#3).
func (s *Server) handleReadyz(w http.ResponseWriter, _ *http.Request) {
	if err := s.pinger.Ping(); err != nil {
		writeError(w, http.StatusServiceUnavailable, "api_unreachable", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, detail string) {
	writeJSON(w, status, map[string]string{"error": code, "detail": detail})
}
