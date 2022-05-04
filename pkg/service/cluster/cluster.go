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
	"muti-kube/pkg/util"
	"muti-kube/pkg/util/logger"
	"os"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
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
	GetClusters(opts ...baseService.OpOption) ([]*cluster.Cluster, *int64, error)
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
func (s *service) GetClusters(opts ...baseService.OpOption) ([]*cluster.Cluster, *int64, error) {
	op := baseService.OpGet(opts...)
	clusterSlice := make([]*cluster.Cluster, 0)
	list, err := s.clustersClient.List(s.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	count := util.ConvertToInt64Ptr(len(list.Items))
	offset, end := baseService.CommonPaginate(list.Items,
		(op.Pagination.Page-1)*op.Pagination.PageSize,
		op.Pagination.PageSize)
	listItem := list.Items[offset:end]
	for _, item := range listItem {
		clientSet, err := s.GetKubernetesClientSet(item.Name)
		if err != nil {
			return nil, nil, err
		}
		nodes, err := s.getClusterNodeInfo(clientSet)
		if err != nil {
			logger.Warn(err)
			clusterSlice = append(clusterSlice, &cluster.Cluster{
				Cluster: item,
				Status:  baseService.SERVICEABNORMAL,
			})
			continue
		}
		var cpuCapacity int64
		var cpuUsage int64
		var memoryCapacity int64
		var memoryUsage int64
		for _, node := range nodes.Items {
			nodeUsage, err := s.getNodeUsage(clientSet, node.Name)
			if err != nil {
				logger.Warn(err)
				continue
			}
			cpuCapacity += node.Status.Capacity.Cpu().ScaledValue(resource.Milli)
			cpuUsage += nodeUsage.Cpu().ScaledValue(resource.Milli)
			// unit M
			memoryUsage += nodeUsage.Memory().ScaledValue(resource.Mega)
			memoryCapacity += node.Status.Capacity.Memory().ScaledValue(resource.Mega)
		}
		if cpuCapacity == 0 {
			cpuCapacity = 1
		}
		if memoryCapacity == 0 {
			memoryCapacity = 1
		}
		clusterSlice = append(clusterSlice, &cluster.Cluster{
			Cluster:           item,
			Status:            baseService.SERVICENORMAL,
			CPUCapacity:       cpuCapacity,
			MemoryCapacity:    memoryCapacity,
			CPUUsage:          cpuUsage,
			MemoryUsage:       memoryUsage,
			CPUUtilisation:    float64(cpuUsage) / float64(cpuCapacity),
			MemoryUtilisation: float64(memoryUsage) / float64(memoryCapacity),
		})
	}
	return clusterSlice, count, nil
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
	clientSet, err := s.GetKubernetesClientSet(clusterName)
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
	clusterItem, err := s.clustersClient.Get(s.ctx, clusterID, metav1.GetOptions{})
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

func (s *service) getNodeUsage(client k8s.Client, nodeName string) (usage v1.ResourceList, err error) {
	metrics, err := client.Metrics().MetricsV1beta1().NodeMetricses().Get(s.ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return metrics.Usage, nil
}

// GetNodeMetric Pass in the cluster ID, node name, and monitoring indicator to obtain the monitoring timing data of the node
func (s *service) GetNodeMetric(metrics []string,
								clusterID string,
								nodeName string,
								) ([]monitoring.Metric, error) {

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
