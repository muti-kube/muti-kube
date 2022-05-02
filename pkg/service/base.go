package service

import (
	"context"
	clusterv1alpha1 "muti-kube/pkg/client/cluster/clientset/versioned/typed/cluster/v1alpha1"
	"muti-kube/pkg/simple/client/monitoring"
	"muti-kube/pkg/simple/client/monitoring/prometheus"
	"reflect"

	"k8s.io/client-go/tools/clientcmd"
)

type base struct {
	ClustersClient clusterv1alpha1.ClusterInterface
	Ctx            context.Context
}

type BaseInterface interface {
	GetPrometheusClient(prometheusURL string) (monitoring.Interface, error)
}

func NewBase() (BaseInterface, error) {
	return newBase()
}

func newBase() (*base, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", DefaultConfigPath)
	if err != nil {
		return nil, err
	}
	clustersClientSet, err := clusterv1alpha1.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &base{
		ClustersClient: clustersClientSet.Clusters(),
		Ctx:            context.Background(),
	}, nil
}

// CommonPaginate Interface memory paging plug-in
func CommonPaginate(x interface{}, offset int, limit int) (int, int) {
	xLen := reflect.ValueOf(x).Len()
	if offset+1 > xLen {
		offset = xLen
	}
	end := offset + limit
	if end > xLen {
		end = xLen
	}
	return offset, end
}

func (bs *base) GetPrometheusClient(prometheusURL string) (monitoring.Interface, error) {
	prometheusOptions := prometheus.NewPrometheusOptions()
	prometheusOptions.Endpoint = prometheusURL
	prometheusClient, err := prometheus.NewPrometheus(prometheusOptions)
	if err != nil {
		return nil, err
	}
	return prometheusClient, nil
}
