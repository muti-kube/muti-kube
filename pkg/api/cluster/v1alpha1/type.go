package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Cluster struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ClusterSpec   `json:"spec"`
	Status            ClusterStatus `json:"status"`
}

// +k8s:deepcopy-gen=false

type ClusterSpec struct {
	KubeConfig    string `json:"kubeconfig"`
	DisplayName   string `json:"displayname"`
	PrometheusURL string `json:"prometheusurl"`
}

// +k8s:deepcopy-gen=false

type ClusterStatus struct {
	CPUCapacity    int64 `json:"cpu_capacity"`
	MemoryCapacity int64 `json:"memory_capacity"`
	CPUUsage       int64 `json:"cpu_usage"`
	MemoryUsage    int64 `json:"memory_usage"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}
