package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/SowmyaRahul/tyk-sre-assignment/internal/k8s"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "path to kubeconfig, leave empty for in-cluster")
	listenAddr := flag.String("address", ":8080", "HTTP server listen address")

	flag.Parse()

	clientset, err := k8s.NewClientset(*kubeconfig)
	if err != nil {
		panic(err)
	}

	version, err := k8s.ServerVersion(clientset)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Connected to Kubernetes %s\n", version)

	if err := startServer(*listenAddr); err != nil {
		panic(err)
	}
}

// launches an HTTP server with defined handlers 
func startServer(listenAddr string) error {
	http.HandleFunc("/healthz", healthHandler)

	fmt.Printf("Server listening on %s\n", listenAddr)

	return http.ListenAndServe(listenAddr, nil)
}

// responds with the health status of the application.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("ok"))
	if err != nil {
		fmt.Println("failed writing to response")
	}
}
