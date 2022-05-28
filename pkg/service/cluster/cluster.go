package cluster

import (
	"context"
	"encoding/json"
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
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	cs     Interface
	csOnce sync.Once
)

type service struct {
	baseService.BaseInterface
	clustersClient clusterv1alpha1.ClusterInterface
	ctx            context.Context
}

type Interface interface {
	GetNodesByClusterID(clusterID string) (*v1.NodeList, error)
	GetKubernetesClientSet(clusterID string) (k8s.Client, error)
	CreateCluster(clusterPost *cluster.Post) (*cluster.Cluster, error)
	GetClusters(opts ...baseService.OpOption) ([]*cluster.Cluster, *int64, error)
	GetNodeUsage(client k8s.Client, nodeName string) (usage v1.ResourceList, err error)
	GetCluster(clusterID string, opts ...baseService.OpOption) (*cluster.Cluster, error)
	GetNodeMetric(metrics []string, clusterID string, nodeName string, start, end time.Time, step time.Duration) ([]monitoring.Metric, error)
}

func NewClusterService() (Interface, error) {
	var err error
	csOnce.Do(func() {
		cs, err = newService()
		if err != nil {
			return
		}
	})
	return cs, nil
}

func newService() (*service, error) {
	base, err := baseService.NewBase()
	if err != nil {
		return nil, err
	}
	return &service{
		clustersClient: base.GetClusterClient(),
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
			continue
		}
		_, err = s.getClusterNodeInfo(clientSet)
		if err != nil {
			logger.Warn(err)
			clusterSlice = append(clusterSlice, &cluster.Cluster{
				Cluster:      item,
				HealthStatus: baseService.Abnormal,
			})
			continue
		}
		versionInfo, err := clientSet.Kubernetes().Discovery().ServerVersion()
		if err != nil {
			continue
		}
		clusterSlice = append(clusterSlice, &cluster.Cluster{
			Cluster:      item,
			Version:      versionInfo.GitVersion,
			HealthStatus: baseService.Normal,
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

func (s *service) GetNodeUsage(client k8s.Client, nodeName string) (usage v1.ResourceList, err error) {
	metrics, err := client.Metrics().MetricsV1beta1().NodeMetricses().Get(s.ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return metrics.Usage, nil
}

// GetNodeMetric Pass in the cluster ID, node name, and monitoring indicator to obtain the monitoring timing data of the node
func (s *service) GetNodeMetric(metrics []string, clusterID string,
	nodeName string, start, end time.Time, step time.Duration) ([]monitoring.Metric, error) {
	clusterData, err := s.GetCluster(clusterID)
	if err != nil {
		return nil, err
	}
	prometheusClient, err := s.BaseInterface.GetPrometheusClient(clusterData.Spec.PrometheusURL)
	if err != nil {
		return nil, err
	}
	var queryOpts []monitoring.QueryOption
	queryOpts = append(queryOpts, monitoring.MeterOption{
		Start: start,
		End:   end,
		Step:  step,
	})
	queryOpts = append(queryOpts, monitoring.NodeOption{NodeName: nodeName})
	metricsValue := prometheusClient.GetNamedMetersOverTime(
		metrics,
		start,
		end,
		step,
		queryOpts,
	)
	return metricsValue, nil
}
