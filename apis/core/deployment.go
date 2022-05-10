package core

import (
	"muti-kube/apis"
	"muti-kube/pkg/consts"
	"muti-kube/pkg/service"
	deploymentService "muti-kube/pkg/service/core"

	"github.com/gin-gonic/gin"
)

type Deployment struct {
	apis.Base
	ds deploymentService.DeploymentInterface
}

func NewDeployment() (*Deployment, error) {
	tmp, err := deploymentService.NewDeployment()
	if err != nil {
		return nil, err
	}
	return &Deployment{
		ds: tmp,
	}, nil
}

func (dc *Deployment)GetDeployments(c *gin.Context){
	pagination := dc.GetPagination(c)
	clusterID := c.Param("clusterID")
	namespace := c.Param("namespace")
	clusters, count, err := dc.ds.GetDeployments(clusterID,namespace,service.WithPagination(pagination))
	if err != nil {
		dc.Error(c, consts.ERRGETDEPLOYMENTS, err, "")
		return
	}
	dc.PageOK(c, clusters, count, pagination, "")
}

