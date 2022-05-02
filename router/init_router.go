package router

import (
	"fmt"
	"muti-kube/router/cluster"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	return baseRouterV1()
}

func baseRouterV1() *gin.Engine {
	r := gin.New()
	addRouter(r.Group(fmt.Sprintf("/api/%s/%s", VERSION, SERVERNAME)))
	return r
}

func addRouter(v1 *gin.RouterGroup) {
	cluster.RegisterClusterRouter(v1)
}
