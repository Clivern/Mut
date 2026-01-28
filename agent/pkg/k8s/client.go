// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package k8s

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Config describes how to reach one Kubernetes cluster.
type Config struct {
	Name           string
	InCluster      bool
	Kubeconfig     string
	KubeconfigData []byte
	Context        string
	QPS            float32
	Burst          int
}

// Cluster is a connected Kubernetes cluster.
type Cluster struct {
	config    Config
	clientset kubernetes.Interface
}

// NewCluster connects to a cluster and verifies the API server is reachable.
func NewCluster(config Config) (*Cluster, error) {
	restConfig, err := LoadConfig(config)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for cluster [%s]: %w", config.Name, err)
	}

	_, err = clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to reach cluster [%s]: %w", config.Name, err)
	}

	return &Cluster{
		config:    config,
		clientset: clientset,
	}, nil
}

// Config returns the configuration used by this cluster.
func (c *Cluster) Config() Config {
	return c.config
}

// LoadConfig builds a REST config from kubeconfig data
func LoadConfig(config Config) (*rest.Config, error) {
	var restConfig *rest.Config
	var err error

	if len(config.KubeconfigData) > 0 {
		restConfig, err = clientcmd.RESTConfigFromKubeConfig(config.KubeconfigData)
	} else if config.InCluster {
		restConfig, err = rest.InClusterConfig()
	} else {
		rules := clientcmd.NewDefaultClientConfigLoadingRules()
		rules.ExplicitPath = config.Kubeconfig

		restConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			rules,
			&clientcmd.ConfigOverrides{CurrentContext: config.Context},
		).ClientConfig()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load config for cluster [%s]: %w", config.Name, err)
	}

	restConfig.QPS = config.QPS
	restConfig.Burst = config.Burst

	return restConfig, nil
}
