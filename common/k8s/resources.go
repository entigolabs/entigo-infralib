package k8s

import (
	"errors"
	"fmt"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetResourcesByGroupVersion(t testing.TestingT, options *k8s.KubectlOptions, groupVersion string) (*metaV1.APIResourceList, error) {
	clientSet, err := k8s.GetKubernetesClientFromOptionsE(t, options)
	if err != nil {
		return nil, err
	}
	return clientSet.Discovery().ServerResourcesForGroupVersion(groupVersion)
}

func WaitUntilResourcesAvailable(
	t testing.TestingT,
	options *k8s.KubectlOptions,
	groupVersion string,
	resources []string,
	retries int,
	sleepBetweenRetries time.Duration,
) error {
	statusMsg := fmt.Sprintf("Wait for resources %s in group %s to be provisioned.", resources, groupVersion)
	message, err := retry.DoWithRetryE(
		t,
		statusMsg,
		retries,
		sleepBetweenRetries,
		func() (string, error) {
			resourceList, err := GetResourcesByGroupVersion(t, options, groupVersion)
			if err != nil {
				return "", err
			}
			for _, resource := range resources {
				if !containsResource(resourceList, resource) {
					return "", errors.New(fmt.Sprintf("Resource %s not found in group %s", resource, groupVersion))
				}
			}
			return "Resources are now available", nil
		},
	)
	if err != nil {
		logger.Logf(t, "Timed out waiting for resources to be provisioned: %s", err)
		return err
	}
	logger.Logf(t, message)
	return nil
}

func containsResource(list *metaV1.APIResourceList, resource string) bool {
	for _, r := range list.APIResources {
		if r.Name == resource {
			return true
		}
	}
	return false
}
