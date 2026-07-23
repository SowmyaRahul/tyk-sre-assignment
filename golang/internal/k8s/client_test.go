package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/version"
	disco "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func TestServerVersion(t *testing.T) {
	t.Run("when the API server reports a version then it is returned", func(t *testing.T) {
		cs := fake.NewSimpleClientset()
		cs.Discovery().(*disco.FakeDiscovery).FakedServerVersion = &version.Info{GitVersion: "1.25.0-fake"}

		v, err := ServerVersion(cs)

		assert.NoError(t, err)
		assert.Equal(t, "1.25.0-fake", v)
	})

	t.Run("when the version info is empty then an empty string is returned", func(t *testing.T) {
		cs := fake.NewSimpleClientset()
		cs.Discovery().(*disco.FakeDiscovery).FakedServerVersion = &version.Info{}

		v, err := ServerVersion(cs)

		assert.NoError(t, err)
		assert.Equal(t, "", v)
	})
}