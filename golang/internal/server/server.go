package server

import (
	"fmt"
	"net/http"

	"k8s.io/client-go/kubernetes"
)

// Server wires HTTP handlers to the Kubernetes client.
type Server struct {
	cs kubernetes.Interface
}


func New(cs kubernetes.Interface) *Server {
	return &Server{cs: cs}
}


func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealthz)
	return mux
}


func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		fmt.Println("failed writing to response")
	}
}
