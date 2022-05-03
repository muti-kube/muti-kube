package cluster

import (
	"muti-kube/apis/cluster"
	"muti-kube/pkg/util/logger"

	"github.com/gin-gonic/gin"
)

func RegisterClusterRouter(v1alpha1 *gin.RouterGroup) {
	clusterApi, err := cluster.NewCluster()
	if err != nil {
		logger.Error(err)
		return
	}
	v1alpha1.GET("/clusters", clusterApi.GetClusters)
	v1alpha1.GET("/clusters/:clusterID", clusterApi.GetCluster)
	v1alpha1.POST("/clusters", clusterApi.CreateCluster)
	v1alpha1.GET("/clusters/:clusterID/nodes/:nodeName/metrics", clusterApi.GetNodeMetrics)
}
