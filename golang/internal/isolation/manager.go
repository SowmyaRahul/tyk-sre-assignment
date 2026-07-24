package isolation

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Manager struct {
	cs kubernetes.Interface
}

func NewManager(cs kubernetes.Interface) *Manager { return &Manager{cs: cs} }


func (m *Manager) Apply(ctx context.Context, req Request, createdBy string) (string, bool, error) {
	id := DeriveID(req)
	pols := BuildPolicies(req, id, createdBy)
	existed := 0
	for i := range pols {
		pol := pols[i]
		_, err := m.cs.NetworkingV1().NetworkPolicies(pol.Namespace).Create(ctx, &pol, metav1.CreateOptions{})
		switch {
		case err == nil:
			// created
		case apierrors.IsAlreadyExists(err):
			existed++
		default:
			return "", false, fmt.Errorf("create policy %s/%s: %w", pol.Namespace, pol.Name, err)
		}
	}
	return id, existed == len(pols), nil
}
