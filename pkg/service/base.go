package service

import (
	"errors"
	"flag"
	"fmt"
	clusterv1alpha1 "muti-kube/pkg/client/cluster/clientset/versioned/typed/cluster/v1alpha1"
	"muti-kube/pkg/simple/client/monitoring"
	"muti-kube/pkg/simple/client/monitoring/prometheus"
	"net/http"
	"path/filepath"
	"reflect"
	"sync"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	bsOnce sync.Once
	bs     *base
)

type base struct {
	ClustersClient clusterv1alpha1.ClusterInterface
}

type BaseInterface interface {
	GetPrometheusClient(prometheusURL string) (monitoring.Interface, error)
	GetClusterClient() clusterv1alpha1.ClusterInterface
}

func NewBase() (BaseInterface, error) {
	bsOnce.Do(func() {
		baseService, err := newBase()
		if err != nil {
			panic(err)
			return
		}
		bs = baseService
	})
	return bs, nil
}

func newBase() (*base, error) {
	var kubeConfig *string
	var err error
	var config *rest.Config
	if home := homedir.HomeDir(); home != "" {
		kubeConfig = flag.String("kubeConfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeConfig = flag.String("kubeConfig", "", "absolute path to the kubeConfig file")
	}
	flag.Parse()
	if config, err = clientcmd.BuildConfigFromFlags("", *kubeConfig); err != nil {
		panic(err.Error())
	}
	clustersClientSet, err := clusterv1alpha1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &base{
		ClustersClient: clustersClientSet.Clusters(),
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
	resp, err := http.Get(prometheusURL)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New(fmt.Sprintf("error connent %s", prometheusURL))
	}
	prometheusOptions := prometheus.NewPrometheusOptions()
	prometheusOptions.Endpoint = prometheusURL
	prometheusClient, err := prometheus.NewPrometheus(prometheusOptions)
	if err != nil {
		return nil, err
	}
	return prometheusClient, nil
}

func (bs *base) GetClusterClient() clusterv1alpha1.ClusterInterface {
	return bs.ClustersClient
}
