package cluster

import (
	"fmt"
	"muti-kube/apis/cluster"

	"github.com/gin-gonic/gin"
)

func RegisterClusterRouter(v1alpha1 *gin.RouterGroup) {
	clusterApi, err := cluster.NewCluster()
	if err != nil {
		fmt.Println(err)
	}
	v1alpha1.GET("/clusters", clusterApi.GetClusters)
	v1alpha1.GET("/clusters/:clusterID", clusterApi.GetCluster)
	v1alpha1.POST("/clusters", clusterApi.CreateCluster)
	v1alpha1.GET("/clusters/:clusterID/nodes/:nodeName/metrics", clusterApi.GetNodeMetrics)
}
