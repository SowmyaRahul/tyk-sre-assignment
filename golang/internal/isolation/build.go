package isolation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	LabelManagedBy   = "app.kubernetes.io/managed-by"
	LabelPartOf      = "app.kubernetes.io/part-of"
	LabelIsolationID = "keights-pod-tracker.io/isolation-id"
	ManagedByValue   = "keights-pod-tracker"
	nsNameLabel      = "kubernetes.io/metadata.name"
)

// Workload identifies a set of pods by namespace + label selector.
type Workload struct {
	Namespace   string            `json:"namespace"`
	PodSelector map[string]string `json:"podSelector"`
}

type Request struct {
	A Workload `json:"a"`
	B Workload `json:"b"`
}

// canonical renders a workload as a stable string (namespace + sorted labels).
func canonical(w Workload) string {
	keys := make([]string, 0, len(w.PodSelector))
	for k := range w.PodSelector {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	sb.WriteString(w.Namespace)
	for _, k := range keys {
		fmt.Fprintf(&sb, ";%s=%s", k, w.PodSelector[k])
	}
	return sb.String()
}

func DeriveID(req Request) string {
	ca, cb := canonical(req.A), canonical(req.B)
	if ca > cb {
		ca, cb = cb, ca
	}
	sum := sha256.Sum256([]byte(ca + "|" + cb))
	return hex.EncodeToString(sum[:])[:10]
}


func allowAllExceptPeer(peer Workload) []networkingv1.NetworkPolicyPeer {
	// 1. Every namespace except the peer's.
	peers := []networkingv1.NetworkPolicyPeer{{
		NamespaceSelector: &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{{
				Key:      nsNameLabel,
				Operator: metav1.LabelSelectorOpNotIn,
				Values:   []string{peer.Namespace},
			}},
		},
	}}
	//  Skipped when the peer is the whole namespace (empty selector).
	if len(peer.PodSelector) > 0 {
		keys := make([]string, 0, len(peer.PodSelector))
		for k := range peer.PodSelector {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		exprs := make([]metav1.LabelSelectorRequirement, 0, len(keys))
		for _, k := range keys {
			exprs = append(exprs, metav1.LabelSelectorRequirement{
				Key:      k,
				Operator: metav1.LabelSelectorOpNotIn,
				Values:   []string{peer.PodSelector[k]},
			})
		}
		peers = append(peers, networkingv1.NetworkPolicyPeer{
			NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{nsNameLabel: peer.Namespace}},
			PodSelector:       &metav1.LabelSelector{MatchExpressions: exprs},
		})
	}
	return peers
}

func policyFor(side string, self, peer Workload, id, createdBy string) networkingv1.NetworkPolicy {
	peers := allowAllExceptPeer(peer)
	pairDesc := fmt.Sprintf("%s/%v <-> %s/%v", self.Namespace, self.PodSelector, peer.Namespace, peer.PodSelector)
	return networkingv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{APIVersion: "networking.k8s.io/v1", Kind: "NetworkPolicy"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kpt-iso-" + id + "-" + side,
			Namespace: self.Namespace,
			Labels: map[string]string{
				LabelManagedBy:   ManagedByValue,
				LabelPartOf:      ManagedByValue,
				LabelIsolationID: id,
			},
			Annotations: map[string]string{
				"keights-pod-tracker.io/created-at": time.Now().UTC().Format(time.RFC3339),
				"keights-pod-tracker.io/pair":       pairDesc,
				"keights-pod-tracker.io/created-by": createdBy,
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: self.PodSelector},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress},
			Ingress:     []networkingv1.NetworkPolicyIngressRule{{From: peers}},
			Egress:      []networkingv1.NetworkPolicyEgressRule{{To: peers}},
		},
	}
}

func BuildPolicies(req Request, id, createdBy string) []networkingv1.NetworkPolicy {
	return []networkingv1.NetworkPolicy{
		policyFor("a", req.A, req.B, id, createdBy),
		policyFor("b", req.B, req.A, id, createdBy),
	}
}
