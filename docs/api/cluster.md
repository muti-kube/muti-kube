# 集群API文档

BASE = `/api/v1alpha1/muti-kube/clusters`

- 获取集群列表

  GET $BASE

  - resp
      - 集群ID: metadata.name

      - 集群名称: spec.displayname

      - 集群创建时间: metadata.creationTimestamp

      - 集群监控地址: spec.prometheusurl

      - 集群管理配置: spec.kubeconfig

      - CPU利用率：cpu_utilisation

      - 内存利用率: memory_utilisation

- 导入集群信息
   
   POST $BASE
  
   - request
     ```json
        {
          "displayname": "集群名称",
          "kubeconfig": "集群配置文件",
          "prometheusurl": "集群监控地址"  
        }    
      ```
