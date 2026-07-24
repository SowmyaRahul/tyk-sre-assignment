package isolation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func sampleReq() Request {
	return Request{
		A: Workload{Namespace: "team-a", PodSelector: map[string]string{"app": "checkout"}},
		B: Workload{Namespace: "team-b", PodSelector: map[string]string{"app": "payments"}},
	}
}

func TestDeriveID(t *testing.T) {
	t.Run("when the same pair is given in either order then the id is identical", func(t *testing.T) {
		r := sampleReq()
		rev := Request{A: r.B, B: r.A}
		assert.Equal(t, DeriveID(r), DeriveID(rev))
	})

	t.Run("when a pair is given then a non-empty id is produced", func(t *testing.T) {
		assert.NotEmpty(t, DeriveID(sampleReq()))
	})
}

func TestBuildPolicies(t *testing.T) {
	// Given the sample pair, when we build policies
	req := sampleReq()
	id := DeriveID(req)
	pols := BuildPolicies(req, id, "api")
	idx := map[string]int{}
	for i, p := range pols {
		idx[p.Namespace] = i
	}
	pa := pols[idx["team-a"]]

	t.Run("when isolating a labeled pair then two symmetric policies with ownership labels are produced", func(t *testing.T) {
		assert.Len(t, pols, 2)
		assert.Equal(t, ManagedByValue, pa.Labels[LabelManagedBy])
		assert.Equal(t, id, pa.Labels[LabelIsolationID])
		assert.NotEmpty(t, pa.Annotations["keights-pod-tracker.io/pair"])
		assert.Equal(t, "checkout", pa.Spec.PodSelector.MatchLabels["app"])
		assert.Len(t, pa.Spec.PolicyTypes, 2)
		assert.Equal(t, "kpt-iso-"+id+"-a", pa.Name)
	})

	t.Run("when a workload is isolated then its egress excludes the peer namespace via NotIn", func(t *testing.T) {
		found := false
		for _, rule := range pa.Spec.Egress {
			for _, peer := range rule.To {
				if peer.NamespaceSelector == nil {
					continue
				}
				for _, exp := range peer.NamespaceSelector.MatchExpressions {
					if exp.Operator == metav1.LabelSelectorOpNotIn {
						for _, v := range exp.Values {
							if v == "team-b" {
								found = true
							}
						}
					}
				}
			}
		}
		assert.True(t, found, "egress should exclude peer namespace team-b via NotIn")
	})
}

func TestBuildPolicies_WholeNamespacePeer(t *testing.T) {
	t.Run("when the peer is a whole namespace then only the exclude-namespace rule is emitted", func(t *testing.T) {
		// Given a peer with an empty selector (the whole namespace)
		req := Request{
			A: Workload{Namespace: "team-a", PodSelector: map[string]string{"app": "checkout"}},
			B: Workload{Namespace: "team-b"},
		}
		// When we build policies
		pols := BuildPolicies(req, DeriveID(req), "api")
		idx := map[string]int{}
		for i, p := range pols {
			idx[p.Namespace] = i
		}
		pa := pols[idx["team-a"]]
		assert.Len(t, pa.Spec.Egress[0].To, 1, "whole-namespace peer needs a single NotIn rule")
	})
}