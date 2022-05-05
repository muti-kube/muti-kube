package cluster

import (
	"context"
	"errors"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestService_GetClusters(t *testing.T) {
	clusterService, err := NewClusterService()
	if err != nil {
		t.Error(err)
		return
	}
	clusters, _, err := clusterService.GetClusters()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(clusters)
}

func TestService_GetKubernetesClientSet(t *testing.T) {
	clusterInterface, err := NewClusterService()
	if err != nil {
		t.Error(err)
		return
	}
	clientSet, err := clusterInterface.GetKubernetesClientSet("cluster-1")
	if err != nil {
		t.Error(err)
		return
	}
	podList, err := clientSet.Kubernetes().CoreV1().Pods("kube-system").
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Error(err)
		return
	}
	if podList == nil || len(podList.Items) == 0 {
		t.Error(errors.New("failure get kubernetes client"))
	}
}

func TestService_GetNodesByClusterID(t *testing.T) {
	clusterInterface, err := NewClusterService()
	if err != nil {
		t.Error(err)
		return
	}
	nodes, err := clusterInterface.GetNodesByClusterID("cluster-1")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(nodes)
}
