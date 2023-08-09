package test

import (
	"context"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"time"
)

func GetProvider(t testing.TestingT, options *k8s.KubectlOptions, providerName string) (*unstructured.Unstructured, error) {
	dynamicClient, err := GetDynamicKubernetesClientFromOptionsE(t, options)
	if err != nil {
		return nil, err
	}
	providerResource := schema.GroupVersionResource{Group: "pkg.crossplane.io", Version: "v1", Resource: "providers"}
	return dynamicClient.Resource(providerResource).Get(context.Background(), providerName, metaV1.GetOptions{})
}

func WaitUntilProviderAvailable(
	t testing.TestingT,
	options *k8s.KubectlOptions,
	providerName string,
	retries int,
	sleepBetweenRetries time.Duration,
) error {
	statusMsg := fmt.Sprintf("Wait for provider %s to be provisioned.", providerName)
	message, err := retry.DoWithRetryE(
		t,
		statusMsg,
		retries,
		sleepBetweenRetries,
		func() (string, error) {
			provider, err := GetProvider(t, options, providerName)
			if err != nil {
				return "", err
			}
			if !IsProviderAvailable(provider) {
				return "", NewProviderNotAvailable(provider)
			}
			return "Provider is now available", nil
		},
	)
	if err != nil {
		logger.Logf(t, "Timed out waiting for provider to be provisioned: %s", err)
		return err
	}
	logger.Logf(t, message)
	return nil
}

func IsProviderAvailable(provider *unstructured.Unstructured) bool {
	status := getProviderStatus(provider)
	return status["Healthy"] == "True" && status["Installed"] == "True"
}

type ProviderNotAvailable struct {
	provider *unstructured.Unstructured
}

// Error is a simple function to return a formatted error message as a string
func (err ProviderNotAvailable) Error() string {
	status := getProviderStatus(err.provider)
	return fmt.Sprintf(
		"Deployment %s is not available, healthy: %s, installed: %s", err.provider.GetName(), status["Healthy"],
		status["Installed"],
	)
}

func NewProviderNotAvailable(provider *unstructured.Unstructured) ProviderNotAvailable {
	return ProviderNotAvailable{provider}
}

func getProviderStatus(provider *unstructured.Unstructured) map[string]string {
	conditions, found, err := unstructured.NestedMap(provider.Object, "status", "conditions")
	if !found || err != nil {
		return nil
	}
	status := make(map[string]string)
	for _, condition := range conditions {
		conditionMap := condition.(map[string]interface{})
		status[conditionMap["type"].(string)] = conditionMap["status"].(string)
	}
	return status
}
