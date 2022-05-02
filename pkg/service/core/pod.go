package core

import (
	"context"
	"muti-kube/pkg/client/k8s"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type podService struct {
	client kubernetes.Interface
}

func NewPodService() *podService {
	client, err := k8s.NewKubernetesClient(&k8s.KubernetesOptions{KubeConfig: "~/.kube/config"})
	if err != nil {
		return nil
	}
	return &podService{
		client: client.Kubernetes(),
	}
}

func (ps *podService) GetPodList() {
	ps.client.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
}
