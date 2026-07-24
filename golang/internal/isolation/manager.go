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

// List returns a map of isolation id -> ["ns/name", ...] for managed policies.
func (m *Manager) List(ctx context.Context) (map[string][]string, error) {
	sel := LabelManagedBy + "=" + ManagedByValue
	pols, err := m.cs.NetworkingV1().NetworkPolicies("").List(ctx, metav1.ListOptions{LabelSelector: sel})
	if err != nil {
		return nil, fmt.Errorf("list policies: %w", err)
	}
	out := map[string][]string{}
	for i := range pols.Items {
		p := &pols.Items[i]
		id := p.Labels[LabelIsolationID]
		out[id] = append(out[id], p.Namespace+"/"+p.Name)
	}
	return out, nil
}


func (m *Manager) Delete(ctx context.Context, id string) (int, error) {
	sel := LabelManagedBy + "=" + ManagedByValue + "," + LabelIsolationID + "=" + id
	pols, err := m.cs.NetworkingV1().NetworkPolicies("").List(ctx, metav1.ListOptions{LabelSelector: sel})
	if err != nil {
		return 0, fmt.Errorf("list for delete: %w", err)
	}
	count := 0
	for i := range pols.Items {
		p := &pols.Items[i]
		if err := m.cs.NetworkingV1().NetworkPolicies(p.Namespace).Delete(ctx, p.Name, metav1.DeleteOptions{}); err != nil {
			return count, fmt.Errorf("delete %s/%s: %w", p.Namespace, p.Name, err)
		}
		count++
	}
	return count, nil
}
