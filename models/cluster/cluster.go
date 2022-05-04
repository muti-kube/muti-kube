package cluster

import (
	"muti-kube/pkg/api/cluster/v1alpha1"

	v1 "k8s.io/api/core/v1"
)

type Cluster struct {
	v1alpha1.Cluster
	Status            string       `json:"status"`
	NodeList          *v1.NodeList `json:"node_list,omitempty"`
	CPUCapacity       int64        `json:"cpu_capacity"`
	MemoryCapacity    int64        `json:"memory_capacity"`
	CPUUsage          int64        `json:"cpu_usage"`
	MemoryUsage       int64        `json:"memory_usage"`
	CPUUtilisation    float64      `json:"cpu_utilisation"`
	MemoryUtilisation float64      `json:"memory_utilisation"`
}

type Post struct {
	DisplayName   string `json:"displayname"`
	KubeConfig    string `json:"kubeconfig"`
	PrometheusURL string `json:"prometheusurl"`
}
