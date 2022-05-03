package cluster

import (
	"muti-kube/apis"
	"muti-kube/models/cluster"
	"muti-kube/pkg/consts"
	"muti-kube/pkg/service"
	clusterService "muti-kube/pkg/service/cluster"
	"strings"

	"github.com/gin-gonic/gin"
)

type Cluster struct {
	apis.Base
	cs clusterService.Interface
}

func NewCluster() (*Cluster, error) {
	tmp, err := clusterService.NewClusterService()
	if err != nil {
		return nil, err
	}
	return &Cluster{
		cs: tmp,
	}, nil
}

func (cc *Cluster) GetClusters(c *gin.Context) {
	pagination := cc.GetPagination(c)
	clusters,count, err := cc.cs.GetClusters(service.WithPagination(pagination))
	if err != nil {
		cc.Error(c,consts.ERRGETCLUSTERS,err,"")
		return
	}
	cc.PageOK(c, clusters, count, pagination, "")
}

func (cc *Cluster) GetCluster(c *gin.Context) {
	clusterID := c.Param("clusterID")
	clusterData, err := cc.cs.GetCluster(clusterID)
	if err != nil {
		cc.Error(c,consts.ERRGETCLUSTER,err,"")
		return
	}
	cc.OK(c, clusterData, "")
}

func (cc *Cluster) CreateCluster(c *gin.Context) {
	clusterPost := &cluster.Post{}
	if err := c.ShouldBindJSON(clusterPost); err != nil {
		cc.Error(c,consts.ERRCREATECLUSTER,err,"")
		return
	}
	clusterData, err := cc.cs.CreateCluster(clusterPost)
	if err != nil {
		cc.Error(c,consts.ERRCREATECLUSTER,err,"")
		return
	}
	cc.OK(c, clusterData, "")
}

func (cc *Cluster) GetNodeMetrics(c *gin.Context) {
	clusterID := c.Param("clusterID")
	nodeName := c.Param("nodeName")
	metrics := c.Query("metrics")
	nodeMetric, err := cc.cs.GetNodeMetric(strings.Split(metrics, ","), clusterID, nodeName)
	if err != nil {
		cc.Error(c,consts.ERRGETNODEMETRICS,err,"")
		return
	}
	cc.OK(c, nodeMetric, "")
}
