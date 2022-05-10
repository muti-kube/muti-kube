package core

import (
	"context"
	baseService "muti-kube/pkg/service"
	"muti-kube/pkg/service/cluster"
	coreModels "muti-kube/models/core"
	"muti-kube/pkg/util"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deploymentService struct {
	baseService.BaseInterface
	ctx            context.Context
	cs             cluster.Interface
}

type DeploymentInterface interface {
	GetDeployments(clusterID string,namespace string,opts ...baseService.OpOption)([]appsv1.Deployment,*int64,error)
}

func NewDeployment() (DeploymentInterface,error) {
	return newDeployment()
} 

func newDeployment() (*deploymentService,error) {
	bs, err := baseService.NewBase()
	if err != nil {
		return nil,err
	}
	clusterService, err := cluster.NewClusterService()
	if err != nil {
		return nil,err
	}
	return &deploymentService{
		ctx: context.Background(),
		BaseInterface: bs,
		cs: clusterService,
	},nil
}

func (ds *deploymentService) GetDeployments(clusterID string,
					namespace string,opts ...baseService.OpOption)([]appsv1.Deployment,*int64,error) {
	op := baseService.OpGet(opts...)
	clientSet, err := ds.cs.GetKubernetesClientSet(clusterID)
	if err != nil {
		return nil,nil,err
	}
	list, err := clientSet.Kubernetes().AppsV1().Deployments(namespace).List(ds.ctx, metav1.ListOptions{})
	if err != nil {
		return nil,nil,err
	}
	count := util.ConvertToInt64Ptr(len(list.Items))
	offset, end := baseService.CommonPaginate(list.Items,
		(op.Pagination.Page-1)*op.Pagination.PageSize,
		op.Pagination.PageSize)
	listItem := list.Items[offset:end]
	return listItem,count,nil
}

func (ds *deploymentService) GetDeployment(clusterID string,
					namespace string,deploymentID string,opts ...baseService.OpOption)(*appsv1.Deployment,error) {
	clientSet, err := ds.cs.GetKubernetesClientSet(clusterID)
	if err != nil {
		return nil,err
	}
	return clientSet.Kubernetes().AppsV1().Deployments(namespace).Get(ds.ctx,deploymentID,metav1.GetOptions{})			
}

func (ds *deploymentService) CreateDeployment(clusterID string,
namespace string,deploymentPost coreModels.DeploymentPost,opts ...baseService.OpOption) {
	clientSet, err := ds.cs.GetKubernetesClientSet(clusterID)
	if err != nil {
		return 
	}
	createDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentPost.Name,
		},
		Spec: deploymentPost.Spec,
	}
	clientSet.Kubernetes().AppsV1().Deployments(namespace).Create(ds.ctx,createDeployment,metav1.CreateOptions{})
}