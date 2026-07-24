package isolation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
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
		assert.NoError(t, err)
		assert.Equal(t, id1, id2)
		assert.True(t, converged2) // second time: converged

		got, err := m.List(ctx)
		assert.NoError(t, err)
		assert.Len(t, got[id1], 2)
	})
}

func TestManager_Delete(t *testing.T) {
	t.Run("when an isolation is deleted then both policies are removed and re-deleting is a no-op", func(t *testing.T) {
		// Given an applied isolation
		cs := fake.NewSimpleClientset()
		m := NewManager(cs)
		ctx := context.Background()
		id, _, _ := m.Apply(ctx, sampleReq(), "api")

		n, err := m.Delete(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, 2, n)
		got, _ := m.List(ctx)
		assert.Len(t, got[id], 0)

		n, err = m.Delete(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, 0, n)
	})
}