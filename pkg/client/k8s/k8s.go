package k8s

import (
	"strings"

	promresourcesclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Client interface {
	Kubernetes() kubernetes.Interface
	Prometheus() promresourcesclient.Interface
	Metrics() metrics.Interface
	Config() *rest.Config
}

type kubernetesClient struct {
	// kubernetes client interface
	k8s kubernetes.Interface
	// discovery client
	discoveryClient discovery.DiscoveryInterface
	// dynamic client
	dynamicClient dynamic.Interface
	master        string
	config        *rest.Config
	prometheus    promresourcesclient.Interface
	apiextensions apiextensionsclient.Interface
	metricsClient metrics.Interface
}

// NewKubernetesClientOrDie creates KubernetesClient and panic if there is an error
func NewKubernetesClientOrDie(options *KubernetesOptions) (client Client) {
	config, err := clientcmd.BuildConfigFromFlags("", options.KubeConfig)
	if err != nil {
		panic(err)
	}

	config.QPS = options.QPS
	config.Burst = options.Burst
	k := &kubernetesClient{
		k8s:             kubernetes.NewForConfigOrDie(config),
		discoveryClient: discovery.NewDiscoveryClientForConfigOrDie(config),
		dynamicClient:   dynamic.NewForConfigOrDie(config),
		master:          config.Host,
		config:          config,
	}

	if options.Master != "" {
		k.master = options.Master
	}
	// The https prefix is automatically added when using sa.
	// But it will not be set automatically when reading from kubeconfig
	// which may cause some problems in the client of other languages.
	if !strings.HasPrefix(k.master, "http://") && !strings.HasPrefix(k.master, "https://") {
		k.master = "https://" + k.master
	}
	return k
}

// NewKubernetesClient creates a KubernetesClient
func NewKubernetesClient(options *KubernetesOptions) (client Client, err error) {
	if options == nil {
		return
	}

	var config *rest.Config
	if config, err = clientcmd.BuildConfigFromFlags("", options.KubeConfig); err != nil {
		return
	}
	config.QPS = options.QPS
	config.Burst = options.Burst

	if client, err = NewKubernetesClientWithConfig(config); err == nil {
		if k8sClient, ok := client.(*kubernetesClient); ok {
			k8sClient.config = config
			k8sClient.master = options.Master
		}
	}
	return
}

// NewKubernetesClientWithConfig creates a k8s client with the rest config
func NewKubernetesClientWithConfig(config *rest.Config) (client Client, err error) {
	if config == nil {
		return
	}

	var k kubernetesClient
	if k.k8s, err = kubernetes.NewForConfig(config); err != nil {
		return
	}

	if k.discoveryClient, err = discovery.NewDiscoveryClientForConfig(config); err != nil {
		return
	}

	if k.dynamicClient, err = dynamic.NewForConfig(config); err != nil {
		return
	}

	if k.apiextensions, err = apiextensionsclient.NewForConfig(config); err != nil {
		return
	}
	if k.prometheus, err = promresourcesclient.NewForConfig(config); err != nil {
		return
	}
	if k.metricsClient, err = metrics.NewForConfig(config); err != nil {
		return
	}
	k.config = config
	client = &k
	return
}

func (k *kubernetesClient) Kubernetes() kubernetes.Interface {
	return k.k8s
}

func (k *kubernetesClient) Config() *rest.Config {
	return k.config
}

func (k *kubernetesClient) Prometheus() promresourcesclient.Interface {
	return k.prometheus
}

func (k *kubernetesClient) Metrics() metrics.Interface {
	return k.metricsClient
}
