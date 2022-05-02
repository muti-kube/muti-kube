package cluster

import (
	"muti-kube/pkg/api/cluster/v1alpha1"

	v1 "k8s.io/api/core/v1"
)

type Cluster struct {
	v1alpha1.Cluster
	NodeList *v1.NodeList
}
