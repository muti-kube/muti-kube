package cluster

import (
	"muti-kube/apis"
	"muti-kube/models/cluster"
	"muti-kube/pkg/consts"
	"muti-kube/pkg/service"
	clusterService "muti-kube/pkg/service/cluster"
	"strconv"
	"strings"
	"time"

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

// GetClusters Obtain the cluster list
func (cc *Cluster) GetClusters(c *gin.Context) {
	pagination := cc.GetPagination(c)
	clusters, count, err := cc.cs.GetClusters(service.WithPagination(pagination))
	if err != nil {
		cc.Error(c, consts.ERRGETCLUSTERS, err, "")
		return
	}
	cc.PageOK(c, clusters, count, pagination, "")
}

// GetCluster Obtain cluster details based on the cluster ID
func (cc *Cluster) GetCluster(c *gin.Context) {
	clusterID := c.Param("clusterID")
	clusterData, err := cc.cs.GetCluster(clusterID)
	if err != nil {
		cc.Error(c, consts.ERRGETCLUSTER, err, "")
		return
	}
	cc.OK(c, clusterData, "")
}

// CreateCluster import the required information to a cluster
func (cc *Cluster) CreateCluster(c *gin.Context) {
	clusterPost := &cluster.Post{}
	if err := c.ShouldBindJSON(clusterPost); err != nil {
		cc.Error(c, consts.ERRCREATECLUSTER, err, "")
		return
	}
	clusterData, err := cc.cs.CreateCluster(clusterPost)
	if err != nil {
		cc.Error(c, consts.ERRCREATECLUSTER, err, "")
		return
	}
	cc.OK(c, clusterData, "")
}

// GetNodeMetrics You can obtain node monitoring indicators based on the cluster ID, node name, and monitoring indicators
func (cc *Cluster) GetNodeMetrics(c *gin.Context) {
	clusterID := c.Param("clusterID")
	nodeName := c.Param("nodeName")
	metrics := c.Query("metrics")
	start := c.Query("start")
	end := c.Query("end")
	step := c.Query("step")
	startTimeStamp, err := strconv.Atoi(start)
	if err != nil {
		cc.Error(c, consts.ERRGETNODEMETRICS, err, "")
		return
	}
	endTimeStamp, err := strconv.Atoi(end)
	if err != nil {
		cc.Error(c, consts.ERRGETNODEMETRICS, err, "")
		return
	}
	stepDuration, err := strconv.Atoi(step)
	if err != nil {
		cc.Error(c, consts.ERRGETNODEMETRICS, err, "")
		return
	}
	nodeMetric, err := cc.cs.GetNodeMetric(strings.Split(metrics, ","),
		clusterID, nodeName,
		time.Unix(int64(startTimeStamp), 0),
		time.Unix(int64(endTimeStamp), 0),
		time.Second*time.Duration(int64(stepDuration)))
	if err != nil {
		cc.Error(c, consts.ERRGETNODEMETRICS, err, "")
		return
	}
	cc.OK(c, nodeMetric, "")
}
