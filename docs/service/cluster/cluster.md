# 集群模块开发文档

测试方法：
    在 deploy/cluster/crd/cluster-cr.yaml 当中填上自己的config信息
    kubectl apply -f deploy/cluster/crd/cluster-crd.yaml
    kubectl apply -f deploy/cluster/crd/cluster-cr.yaml

集群逻辑代码位置为 pkg/service/cluster



