package k8s

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/testing"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// GetDynamicKubernetesClientFromOptionsE returns a dynamic Kubernetes API client given a configured KubectlOptions object.
func GetDynamicKubernetesClientFromOptionsE(t testing.TestingT, options *k8s.KubectlOptions) (dynamic.Interface, error) {
	var err error
	var config *rest.Config

	if options.InClusterAuth {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		logger.Log(t, "Configuring Kubernetes client to use the in-cluster serviceaccount token")
	} else {
		kubeConfigPath, err := options.GetConfigPath(t)
		if err != nil {
			return nil, err
		}
		logger.Logf(t, "Configuring Kubernetes client using config file %s with context %s", kubeConfigPath, options.ContextName)
		// Load API config (instead of more low level ClientConfig)
		config, err = k8s.LoadApiClientConfigE(kubeConfigPath, options.ContextName)
		if err != nil {
			logger.Logf(t, "Error loading api client config, falling back to in-cluster authentication via serviceaccount token: %s", err)
			config, err = rest.InClusterConfig()
			if err != nil {
				return nil, err
			}
			logger.Log(t, "Configuring Kubernetes client to use the in-cluster serviceaccount token")
		}
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return dynamicClient, nil
}
