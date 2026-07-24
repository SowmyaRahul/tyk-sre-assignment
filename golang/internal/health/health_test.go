package health

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func dep(name string, gen, observed int64, desired *int32, ready int32) appsv1.Deployment {
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: name, Generation: gen},
		Spec:       appsv1.DeploymentSpec{Replicas: desired},
		Status:     appsv1.DeploymentStatus{ObservedGeneration: observed, ReadyReplicas: ready},
	}
}

func ptr(i int32) *int32 { return &i }

func TestCompute(t *testing.T) {
	t.Run("when ready replicas meet desired and generation is current then it is healthy", func(t *testing.T) {
		list := &appsv1.DeploymentList{Items: []appsv1.Deployment{dep("web", 1, 1, ptr(3), 3)}}
		
		r := Compute(list)
		
		assert.True(t, r.Deployments[0].Healthy)
		assert.Equal(t, 1, r.Summary.Healthy)
	})

	t.Run("when ready replicas are fewer than desired then it is unhealthy", func(t *testing.T) {
		list := &appsv1.DeploymentList{Items: []appsv1.Deployment{dep("web", 1, 1, ptr(3), 1)}}
		r := Compute(list)
		assert.False(t, r.Deployments[0].Healthy)
		assert.Equal(t, 1, r.Summary.Unhealthy)
	})

	t.Run("when observed generation lags the spec then it is unhealthy", func(t *testing.T) {
		list := &appsv1.DeploymentList{Items: []appsv1.Deployment{dep("web", 2, 1, ptr(1), 1)}}
		r := Compute(list)
		assert.False(t, r.Deployments[0].Healthy)
	})

	t.Run("when spec.replicas is nil then desired defaults to 1", func(t *testing.T) {
		list := &appsv1.DeploymentList{Items: []appsv1.Deployment{dep("web", 1, 1, nil, 1)}}
		r := Compute(list)
		assert.Equal(t, int32(1), r.Deployments[0].Desired)
		assert.True(t, r.Deployments[0].Healthy)
	})

	t.Run("when the list is empty then the report is empty", func(t *testing.T) {
		r := Compute(&appsv1.DeploymentList{})
		// Then totals are zero and there are no rows
		assert.Equal(t, 0, r.Summary.Total)
		assert.Len(t, r.Deployments, 0)
	})

	t.Run("when there is a mix then the summary counts both healthy and unhealthy", func(t *testing.T) {
		list := &appsv1.DeploymentList{Items: []appsv1.Deployment{
			dep("ok", 1, 1, ptr(1), 1),
			dep("bad", 1, 1, ptr(2), 0),
		}}
		r := Compute(list)
		assert.Equal(t, 2, r.Summary.Total)
		assert.Equal(t, 1, r.Summary.Healthy)
		assert.Equal(t, 1, r.Summary.Unhealthy)
	})
}