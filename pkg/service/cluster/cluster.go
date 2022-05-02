package cluster

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"muti-kube/models/cluster"
	"muti-kube/pkg/api/cluster/v1alpha1"
	clusterv1alpha1 "muti-kube/pkg/client/cluster/clientset/versioned/typed/cluster/v1alpha1"
	"muti-kube/pkg/client/k8s"
	baseService "muti-kube/pkg/service"
	"muti-kube/pkg/simple/client/monitoring"
	"os"
	"path/filepath"
	"time"

	"k8s.io/apimachinery/pkg/util/rand"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type service struct {
	baseService.BaseInterface
	clustersClient clusterv1alpha1.ClusterInterface
	ctx            context.Context
}

type Interface interface {
	GetClusters(opts ...baseService.OpOption) ([]*cluster.Cluster, error)
	GetCluster(clusterID string, opts ...baseService.OpOption) (*cluster.Cluster, error)
	CreateCluster(clusterPost *cluster.Post) (*cluster.Cluster, error)
	GetKubernetesClientSet(clusterID string) (k8s.Client, error)
	GetNodesByClusterID(clusterID string) (*v1.NodeList, error)
	GetNodeMetric(metrics []string, clusterID string, nodeName string) ([]monitoring.Metric, error)
}

func NewClusterService() (Interface, error) {
	return newService()
}

func newService() (*service, error) {
	var kubeconfig *string
	var err error
	var config *rest.Config
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "kubeconfig absolute path")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "kubeconfig absolute path")
	}
	flag.Parse()

	// 首先使用 inCluster 模式(需要去配置对应的 RBAC 权限，默认的sa是default->是没有获取deployments的List权限)
	if config, err = rest.InClusterConfig(); err != nil {
		// 使用 KubeConfig 文件创建集群配置 Config 对象
		if config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig); err != nil {
			panic(err.Error())
		}
	}
	clustersClientSet, err := clusterv1alpha1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	base, err := baseService.NewBase()
	if err != nil {
		return nil, err
	}
	return &service{
		clustersClient: clustersClientSet.Clusters(),
		ctx:            context.Background(),
		BaseInterface:  base,
	}, nil
}

// GetClusters Obtain the cluster list and brief information about the nodes in the cluster
func (s *service) GetClusters(opts ...baseService.OpOption) ([]*cluster.Cluster, error) {
	clusterSlice := make([]*cluster.Cluster, 0)
	list, err := s.clustersClient.List(s.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, item := range list.Items {
		clientSet, err := s.GetKubernetesClientSet(item.Name)
		if err != nil {
			return nil, err
		}
		nodes, err := s.getClusterNodeInfo(clientSet)
		if err != nil {
			return nil, err
		}
		clusterSlice = append(clusterSlice, &cluster.Cluster{
			Cluster:  item,
			NodeList: nodes,
		})
	}
	return clusterSlice, nil
}

func (s *service) GetCluster(clusterID string, opts ...baseService.OpOption) (*cluster.Cluster, error) {
	clusterData, err := s.clustersClient.Get(s.ctx, clusterID, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	clientSet, err := s.GetKubernetesClientSet(clusterData.Name)
	if err != nil {
		return nil, err
	}
	nodes, err := s.getClusterNodeInfo(clientSet)
	if err != nil {
		return nil, err
	}
	return &cluster.Cluster{
		Cluster:  *clusterData,
		NodeList: nodes,
	}, nil
}

func (s *service) CreateCluster(clusterPost *cluster.Post) (*cluster.Cluster, error) {
	randomStr := rand.String(6)
	clusterName := fmt.Sprintf("cluster-%s", randomStr)
	clusterData, err := s.clustersClient.Create(s.ctx, &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
		Spec: v1alpha1.ClusterSpec{
			DisplayName:   clusterPost.DisplayName,
			KubeConfig:    clusterPost.KubeConfig,
			PrometheusURL: clusterPost.PrometheusURL,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	clientSet, err := s.GetKubernetesClientSet(clusterData.Name)
	if err != nil {
		return nil, err
	}
	nodes, err := s.getClusterNodeInfo(clientSet)
	if err != nil {
		return nil, err
	}
	return &cluster.Cluster{
		Cluster:  *clusterData,
		NodeList: nodes,
	}, nil
}

// GetKubernetesClientSet Get the kubernetes native clientSet
func (s *service) GetKubernetesClientSet(clusterID string) (k8s.Client, error) {
	_, err := s.clustersClient.List(s.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	clusterItem, err := s.clustersClient.Get(s.ctx, "cluster-1", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	tmpFile, err := ioutil.TempFile(os.TempDir(), baseService.TempClusterFilePrefix)
	if err != nil {
		return nil, err
	}
	_, err = tmpFile.Write([]byte(clusterItem.Spec.KubeConfig))
	if err != nil {
		return nil, err
	}
	config, err := clientcmd.BuildConfigFromFlags("", tmpFile.Name())
	if err != nil {
		return nil, err
	}
	return k8s.NewKubernetesClientWithConfig(config)
}

// GetNodesByClusterID If the cluster ID is passed in, information about the nodes in the current cluster is returned
func (s *service) GetNodesByClusterID(clusterID string) (*v1.NodeList, error) {
	clientSet, err := s.GetKubernetesClientSet(clusterID)
	if err != nil {
		return nil, err
	}
	return s.getClusterNodeInfo(clientSet)
}

func (s *service) CordonNodeBy(clusterID string, nodeName string) error {
	payload := []cluster.PatchStringValue{{
		Op:    "replace",
		Path:  "/spec/unschedulable",
		Value: true,
	},
	}
	clientSet, err := s.GetKubernetesClientSet(clusterID)
	if err != nil {
		return err
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = clientSet.Kubernetes().
		CoreV1().Nodes().Patch(
		s.ctx, nodeName,
		types.MergePatchType,
		payloadBytes, metav1.PatchOptions{},
	)
	return err
}

// getClusterNodeInfo Gets the cluster node where the incoming client resides
func (s *service) getClusterNodeInfo(client k8s.Client) (*v1.NodeList, error) {
	return client.Kubernetes().CoreV1().Nodes().List(s.ctx, metav1.ListOptions{})
}

func (s *service) GetNodeMetric(metrics []string, clusterID string, nodeName string) ([]monitoring.Metric, error) {
	clusterData, err := s.GetCluster(clusterID)
	if err != nil {
		return nil, err
	}
	prometheusClient, err := s.BaseInterface.GetPrometheusClient(clusterData.Spec.PrometheusURL)
	if err != nil {
		return nil, err
	}
	var queryOpts []monitoring.QueryOption
	startTime := time.Now().Add(-time.Hour * 3)
	endTime := time.Now()
	stepTime := time.Second
	queryOpts = append(queryOpts, monitoring.MeterOption{
		Start: startTime,
		End:   endTime,
		Step:  stepTime,
	})
	queryOpts = append(queryOpts, monitoring.NodeOption{NodeName: nodeName})
	metricsValue := prometheusClient.GetNamedMetersOverTime(
		metrics,
		startTime,
		endTime,
		stepTime,
		queryOpts,
	)
	return metricsValue, nil
}
