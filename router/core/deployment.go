package core

import (
	coreService "muti-kube/apis/core"
	"muti-kube/pkg/util/logger"

	"github.com/gin-gonic/gin"
)

func RegisterDeploymentRouter(v1alpha1 *gin.RouterGroup) {
	deploymentApi, err := coreService.NewDeployment()
	if err != nil {
		logger.Error(err)
		return
	}
	v1alpha1.GET("/clusters/:clusterID/namespaces/:namespace/deployments", deploymentApi.GetDeployments)
}