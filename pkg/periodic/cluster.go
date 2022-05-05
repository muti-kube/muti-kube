package periodic

import (
	"context"
	"fmt"
	"muti-kube/models/common"
	clusterv1alpha1 "muti-kube/pkg/client/cluster/clientset/versioned/typed/cluster/v1alpha1"
	"muti-kube/pkg/consts"
	baseService "muti-kube/pkg/service"
	clusterService "muti-kube/pkg/service/cluster"
	"muti-kube/pkg/util/logger"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

type ClusterPeriodic interface {
	Start()
}

type clusterPeriodic struct {
	bs             baseService.BaseInterface
	cs             clusterService.Interface
	clustersClient clusterv1alpha1.ClusterInterface
}

func NewTicketPeriodic() (ClusterPeriodic, error) {
	cluster, err := clusterService.NewClusterService()
	if err != nil {
		return nil, err
	}
	base, err := baseService.NewBase()
	if err != nil {
		return nil, err
	}
	return &clusterPeriodic{
		cs: cluster,
		bs: base,
	}, nil
}

func (cp *clusterPeriodic) Start() {
	go wait.Forever(cp.updateClusterResourceStatus, time.Minute)
}

func (cp *clusterPeriodic) updateClusterResourceStatus() {
	clusters, _, err := cp.cs.GetClusters(baseService.WithPagination(&common.Pagination{
		Page:     consts.DEFAULT_PAGE,
		PageSize: consts.DEFAULT_PAGE_SIZE,
	}))
	if err != nil {
		logger.Error(err)
		return
	}
	var cpuCapacity int64
	var cpuUsage int64
	var memoryCapacity int64
	var memoryUsage int64
	clusterClient := cp.bs.GetClusterClient()
	for _, cluster := range clusters {
		nodeList, err := cp.cs.GetNodesByClusterID(cluster.Name)
		if err != nil {
			continue
		}
		clientSet, err := cp.cs.GetKubernetesClientSet(cluster.Name)
		if err != nil {
			continue
		}
		for _, node := range nodeList.Items {
			nodeUsage, err := cp.cs.GetNodeUsage(clientSet, node.Name)
			if err != nil {
				logger.Warn(fmt.Sprintf("cluster: %s ", cluster.Name), err)
				continue
			}
			cpuCapacity += node.Status.Capacity.Cpu().ScaledValue(resource.Milli)
			cpuUsage += nodeUsage.Cpu().ScaledValue(resource.Milli)
			// unit M
			memoryUsage += nodeUsage.Memory().ScaledValue(resource.Mega)
			memoryCapacity += node.Status.Capacity.Memory().ScaledValue(resource.Mega)
		}
		patchData := []byte(
			fmt.Sprintf(
				`{"status": {
					"cpu_usage": %d,
					"memory_usage":%d,
					"cpu_capacity":%d,
					"memory_capacity":%d}}`,
				cpuUsage, memoryUsage, cpuCapacity, memoryCapacity))
		_, err = clusterClient.Patch(
			context.Background(),
			cluster.Name,
			types.MergePatchType,
			patchData, metav1.PatchOptions{},
		)
		if err != nil {
			logger.Warn(err)
		}
	}
}
