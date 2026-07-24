package health

import appsv1 "k8s.io/api/apps/v1"


type DeploymentStatus struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Desired   int32  `json:"desired"`
	Ready     int32  `json:"ready"`
	Available int32  `json:"available"`
	Updated   int32  `json:"updated"`
	Healthy   bool   `json:"healthy"`
}


type Report struct {
	Summary struct {
		Total     int `json:"total"`
		Healthy   int `json:"healthy"`
		Unhealthy int `json:"unhealthy"`
	} `json:"summary"`
	Deployments []DeploymentStatus `json:"deployments"`
}


func Compute(list *appsv1.DeploymentList) Report {
	var r Report
	r.Deployments = make([]DeploymentStatus, 0, len(list.Items))
	for i := range list.Items {
		d := &list.Items[i]
		desired := int32(1) 
		if d.Spec.Replicas != nil {
			desired = *d.Spec.Replicas
		}
		healthy := d.Status.ReadyReplicas >= desired &&
			d.Status.ObservedGeneration >= d.Generation
		r.Deployments = append(r.Deployments, DeploymentStatus{
			Namespace: d.Namespace,
			Name:      d.Name,
			Desired:   desired,
			Ready:     d.Status.ReadyReplicas,
			Available: d.Status.AvailableReplicas,
			Updated:   d.Status.UpdatedReplicas,
			Healthy:   healthy,
		})
		r.Summary.Total++
		if healthy {
			r.Summary.Healthy++
		} else {
			r.Summary.Unhealthy++
		}
	}
	return r
}
