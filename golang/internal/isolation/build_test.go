package isolation

import (
	"testing"

	"github.com/stretchr/testify/assert"
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