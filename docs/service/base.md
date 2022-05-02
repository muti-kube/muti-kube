# BaseService

获取kubernetes的原生client

代码位置 pkg/service/base.go
    - baseInterface.GetKubernetesClientSet({ClusterID})
        列表选择集群后，传到后端接口为ClusterID，就是Cluster的CRD的name,
        通过BaseService获取kubernetes的原生Client
