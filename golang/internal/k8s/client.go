package k8s

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// NewClientset builds a clientset from the given kubeconfig path. 
func NewClientset(kubeconfig string) (kubernetes.Interface, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("build kube config: %w", err)
	}
	return kubernetes.NewForConfig(cfg)
}

// API server's GitVersion. 
func ServerVersion(cs kubernetes.Interface) (string, error) {
	v, err := cs.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return v.String(), nil
}

// Pinger reports whether the API server is reachable.
type Pinger interface {
	Ping() error
}

type pinger struct{ cs kubernetes.Interface }

// NewPinger returns a Pinger backed by the given clientset.
func NewPinger(cs kubernetes.Interface) Pinger { return &pinger{cs: cs} }


func (p *pinger) Ping() error {
	_, err := p.cs.Discovery().ServerVersion()
	return err
}
