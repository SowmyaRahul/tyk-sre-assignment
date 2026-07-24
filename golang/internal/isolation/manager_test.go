package isolation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestManager_Apply(t *testing.T) {
	t.Run("when the same pair is applied twice then it converges with no duplicate policies", func(t *testing.T) {
		cs := fake.NewSimpleClientset()
		m := NewManager(cs)
		ctx := context.Background()

		// When we apply the pair the first time
		id1, converged1, err := m.Apply(ctx, sampleReq(), "api")
		assert.NoError(t, err)
		assert.False(t, converged1)

		// When we apply the same pair again
		id2, converged2, err := m.Apply(ctx, sampleReq(), "api")
		// Then it converges to the same id
		assert.NoError(t, err)
		assert.Equal(t, id1, id2)
		assert.True(t, converged2)

		pols, err := cs.NetworkingV1().NetworkPolicies("").List(ctx, metav1.ListOptions{})
		assert.NoError(t, err)
		assert.Len(t, pols.Items, 2)
	})
}