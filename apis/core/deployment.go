package core

import (
	"fmt"
	"muti-kube/apis"
	"muti-kube/models/core"
	"muti-kube/pkg/consts"
	"muti-kube/pkg/service"
	deploymentService "muti-kube/pkg/service/core"

	"github.com/gin-gonic/gin"
)

type Deployment struct {
	apis.Base
	ds deploymentService.DeploymentInterface
	deploymentActionFunc map[string]func(dc *Deployment,c *gin.Context)
}

func dryRunDeployment(dc *Deployment,c *gin.Context)  {
	clusterID := c.Param("clusterID")
	namespace := c.Param("namespace")
	deploymentPost := core.DeploymentPost{}
	if err := c.ShouldBindJSON(&deploymentPost); err != nil {
		dc.Error(c, consts.ERRCREATECLUSTER, err, "")
		return
	}
	deployment, err := dc.ds.DryRunDeployment(clusterID, namespace, deploymentPost)
	if err != nil {
		dc.Error(c, consts.ErrorGetDeployments, err, "")
		return
	}
	dc.OK(c, deployment, "dry-run deployment success")
}

func createDeployment(dc *Deployment,c *gin.Context)  {
	clusterID := c.Param("clusterID")
	namespace := c.Param("namespace")
	deploymentPost := core.DeploymentPost{}
	if err := c.ShouldBindJSON(&deploymentPost); err != nil {
		dc.Error(c, consts.ERRCREATECLUSTER, err, "")
		return
	}
	deployment, err := dc.ds.CreateDeployment(clusterID,namespace,deploymentPost)
	if err != nil {
		dc.Error(c, consts.ErrorCreateDeployment, err, "")
		return
	}
	dc.OK(c, deployment, "dry-run deployment success")
}

func scaleReplicasDeployment(dc *Deployment,c *gin.Context)  {
	clusterID := c.Param("clusterID")
	namespace := c.Param("namespace")
	deploymentPost := core.DeploymentPost{}
	if err := c.ShouldBindJSON(&deploymentPost); err != nil {
		dc.Error(c, consts.ErrorScaleReplicasDeployment, err, "")
		return
	}
	deployment, err := dc.ds.CreateDeployment(clusterID,namespace,deploymentPost)
	if err != nil {
		dc.Error(c, consts.ErrorScaleReplicasDeployment, err, "")
		return
	}
	dc.OK(c, deployment, fmt.Sprintf("relicas %d deployment success",deploymentPost.Replicas))
}

func newDeploymentActionFunc()  map[string]func(dc *Deployment,c *gin.Context) {
	return map[string]func(dc *Deployment,c *gin.Context){
		apis.DryRunAction: dryRunDeployment,
		apis.CreateAction: createDeployment,
		apis.ScaleReplicasAction: scaleReplicasDeployment,
	}
}

func NewDeployment() (*Deployment, error) {
	tmp, err := deploymentService.NewDeployment()
	if err != nil {
		return nil, err
	}
	return &Deployment{
		ds: tmp,
		deploymentActionFunc: newDeploymentActionFunc(),
	}, nil
}

func (dc *Deployment) GetDeployments(c *gin.Context) {
	pagination := dc.GetPagination(c)
	clusterID := c.Param("clusterID")
	namespace := c.Param("namespace")
	clusters, count, err := dc.ds.GetDeployments(clusterID, namespace, service.WithPagination(pagination))
	if err != nil {
		dc.Error(c, consts.ErrorGetDeployments, err, "")
		return
	}
	dc.PageOK(c, clusters, count, pagination, "")
}

func (dc *Deployment) DeploymentAction(c *gin.Context) {
	action := c.DefaultQuery("action", "create")
	actionFunc := dc.deploymentActionFunc[action]
	actionFunc(dc,c)
}

func (dc *Deployment) DeleteDeployment(c *gin.Context) {
	clusterID := c.Param("clusterID")
	namespace := c.Param("namespace")
	deploymentID := c.Param("deploymentID")
	err := dc.ds.DeleteDeployment(clusterID, namespace, deploymentID)
	if err != nil {
		dc.Error(c, consts.ErrorDeleteDeployment, err, "")
		return
	}
	dc.OK(c,nil,fmt.Sprintf("delete deployment %s success",deploymentID))
}

func (dc *Deployment) GetDeployment(c *gin.Context)  {
	clusterID := c.Param("clusterID")
	namespace := c.Param("namespace")
	deploymentID := c.Param("deploymentID")
	deployment, err := dc.ds.GetDeployment(clusterID, namespace, deploymentID)
	if err != nil {
		dc.Error(c, consts.ErrorGetDeployment, err, "")
		return
	}
	dc.OK(c,deployment,fmt.Sprintf("get deployment %s success",deploymentID))
}
