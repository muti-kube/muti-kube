package core

import appsv1 "k8s.io/api/apps/v1"

type DeploymentPost struct {
	appsv1.Deployment `json:",inline"`
	Replicas int32 `json:"replicas"`
}
