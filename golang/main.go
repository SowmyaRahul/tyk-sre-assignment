package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/SowmyaRahul/tyk-sre-assignment/internal/k8s"
	"github.com/SowmyaRahul/tyk-sre-assignment/internal/server"
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

	srv := server.New(clientset, k8s.NewPinger(clientset))

	fmt.Printf("Server listening on %s\n", *listenAddr)
	if err := http.ListenAndServe(*listenAddr, srv.Handler()); err != nil {
		panic(err)
	}
}
