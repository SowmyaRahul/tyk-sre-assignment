package isolation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

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
