package util

import "k8s.io/client-go/rest"

func IsInCluster() bool {
	config, err := rest.InClusterConfig()
	if err != nil || config == nil {
		return false
	}
	return true
}
