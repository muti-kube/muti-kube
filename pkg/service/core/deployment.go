package core

import (
	"context"
	coreModels "muti-kube/models/core"
	baseService "muti-kube/pkg/service"
	"muti-kube/pkg/service/cluster"
	"muti-kube/pkg/util"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deploymentService struct {
	baseService.BaseInterface
	ctx context.Context
	cs  cluster.Interface
}

type DeploymentInterface interface {
	DeleteDeployment(clusterID string, namespace string, deploymentID string) error
	ScaleDeployment(clusterID string, namespace string, deploymentID string,replicas int32) error
	GetDeployments(clusterID string, namespace string, opts ...baseService.OpOption) ([]appsv1.Deployment, *int64, error)
	GetDeployment(clusterID string, namespace string, deploymentID string, opts ...baseService.OpOption) (*appsv1.Deployment, error)
	CreateDeployment(clusterID string, namespace string, deploymentPost coreModels.DeploymentPost, opts ...baseService.OpOption) (*appsv1.Deployment,error)
	DryRunDeployment(clusterID string, namespace string, deploymentPost coreModels.DeploymentPost, opts ...baseService.OpOption) (*appsv1.Deployment, error)
}

func NewDeployment() (DeploymentInterface, error) {
	return newDeployment()
}

func newDeployment() (*deploymentService, error) {
	bs, err := baseService.NewBase()
	if err != nil {
		return nil, err
	}
	clusterService, err := cluster.NewClusterService()
	if err != nil {
		return nil, err
	}
	return &deploymentService{
		ctx:           context.Background(),
		BaseInterface: bs,
		cs:            clusterService,
	}, nil
}

func (ds *deploymentService) GetDeployments(clusterID string,
	namespace string, opts ...baseService.OpOption) ([]appsv1.Deployment, *int64, error) {
	op := baseService.OpGet(opts...)
	clientSet, err := ds.cs.GetKubernetesClientSet(clusterID)
	if err != nil {
		return nil, nil, err
	}
	list, err := clientSet.Kubernetes().AppsV1().Deployments(namespace).List(ds.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	count := util.ConvertToInt64Ptr(len(list.Items))
	offset, end := baseService.CommonPaginate(list.Items,
		(op.Pagination.Page-1)*op.Pagination.PageSize,
		op.Pagination.PageSize)
	listItem := list.Items[offset:end]
	return listItem, count, nil
}

func (ds *deploymentService) GetDeployment(
	clusterID string,
	namespace string,
	deploymentID string,
	opts ...baseService.OpOption,
) (*appsv1.Deployment, error) {
	clientSet, err := ds.cs.GetKubernetesClientSet(clusterID)
	if err != nil {
		return nil, err
	}
	return clientSet.Kubernetes().AppsV1().Deployments(namespace).Get(ds.ctx, deploymentID, metav1.GetOptions{})
}

func (ds *deploymentService) DryRunDeployment(
	clusterID string,
	namespace string,
	deploymentPost coreModels.DeploymentPost,
	opts ...baseService.OpOption,
) (*appsv1.Deployment, error) {
	clientSet, err := ds.cs.GetKubernetesClientSet(clusterID)
	if err != nil {
		return nil, nil
	}
	createDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentPost.Name,
			Namespace: namespace,
		},
		Spec: deploymentPost.Spec,
	}
	return clientSet.Kubernetes().AppsV1().Deployments("default").Create(
		ds.ctx,
		createDeployment,
		metav1.CreateOptions{
			DryRun: []string{metav1.DryRunAll},
		},
	)
}

func (ds *deploymentService) CreateDeployment(
	clusterID string,
	namespace string,
	deploymentPost coreModels.DeploymentPost,
	opts ...baseService.OpOption) (*appsv1.Deployment,error) {
	clientSet, err := ds.cs.GetKubernetesClientSet(clusterID)
	if err != nil {
		return nil,err
	}
	createDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentPost.Name,
		},
		Spec: deploymentPost.Spec,
	}
	deployment, err := clientSet.Kubernetes().AppsV1().Deployments(namespace).Create(ds.ctx, createDeployment, metav1.CreateOptions{})
	if err != nil {
		return nil,err
	}
	return deployment,nil
}

func (ds *deploymentService) ScaleDeployment(
	clusterID string,
	namespace string,
	deploymentID string,
	replicas int32) error {
	clientSet, err := ds.cs.GetKubernetesClientSet(clusterID)
	if err != nil {
		return err
	}
	scale,err := clientSet.Kubernetes().AppsV1().Deployments(namespace).GetScale(context.TODO(),deploymentID,metav1.GetOptions{})
	if err != nil {
		return err
	}
	scale.Spec.Replicas = replicas
	_, err = clientSet.Kubernetes().AppsV1().Deployments(namespace).UpdateScale(ds.ctx, deploymentID, scale, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (ds *deploymentService) DeleteDeployment(
	clusterID string,
	namespace string,
	deploymentID string,
	) error {
	clientSet, err := ds.cs.GetKubernetesClientSet(clusterID)
	if err != nil {
		return err
	}
	err = clientSet.Kubernetes().AppsV1().Deployments(namespace).Delete(ds.ctx,deploymentID,metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (ds *deploymentService) UpdateDeployment(
	clusterID string,
	namespace string,
	deploymentID string,
	deploymentPost coreModels.DeploymentPost,
)(*appsv1.Deployment,error){
	clientSet, err := ds.cs.GetKubernetesClientSet(clusterID)
	if err != nil {
		return nil,err
	}
	updateDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentID,
		},
		Spec: deploymentPost.Spec,
	}
	deployment,err := clientSet.Kubernetes().AppsV1().Deployments(namespace).Update(ds.ctx,updateDeployment,metav1.UpdateOptions{})
	if err != nil {
		return nil,err
	}
	return deployment,nil
}
