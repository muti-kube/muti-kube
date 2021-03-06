/*
Copyright The kube-cloud Authors.
*/
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "muti-kube/pkg/client/cluster/clientset/versioned/typed/cluster/v1alpha1"

	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeCrdV1alpha1 struct {
	*testing.Fake
}

func (c *FakeCrdV1alpha1) Clusters() v1alpha1.ClusterInterface {
	return &FakeClusters{c}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeCrdV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
