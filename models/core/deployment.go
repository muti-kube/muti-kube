package core

import appsv1 "k8s.io/api/apps/v1"

type DeploymentPost struct {
	Name string `json:"name"`
	Labels map[string]string `json:"labels"`
	Spec appsv1.DeploymentSpec `json:"spec"`
}