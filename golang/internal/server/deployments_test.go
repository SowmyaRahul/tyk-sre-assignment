package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SowmyaRahul/tyk-sre-assignment/internal/health"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDeployments(t *testing.T) {
	t.Run("when GET /deployments then it returns a 200 health report that flags a degraded deployment", func(t *testing.T) {
		rep := int32(2)
		cs := fake.NewSimpleClientset(&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "web", Generation: 1},
			Spec:       appsv1.DeploymentSpec{Replicas: &rep},
			Status:     appsv1.DeploymentStatus{ObservedGeneration: 1, ReadyReplicas: 1},
		})
		srv := New(cs, stubPinger{})

		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/deployments", nil))

		assert.Equal(t, http.StatusOK, rec.Code)
		var r health.Report
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &r))
		assert.Equal(t, 1, r.Summary.Total)
		assert.Equal(t, 1, r.Summary.Unhealthy)
	})

	t.Run("when a non-GET method is used then it returns 405", func(t *testing.T) {
		srv := New(fake.NewSimpleClientset(), stubPinger{})
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/deployments", nil))
		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	})
}