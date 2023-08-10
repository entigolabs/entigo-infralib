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
	k8sYaml "k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"time"
)

func WaitUntilProviderAvailable(t testing.TestingT, options *k8s.KubectlOptions, name string, retries int, sleepBetweenRetries time.Duration) (*unstructured.Unstructured, error) {
	resource := schema.GroupVersionResource{Group: "pkg.crossplane.io", Version: "v1", Resource: "providers"}
	availability := defaultObjectAvailability(name, resource)
	availability.isAvailable = isProviderAvailable
	availability.objectError = NewProviderNotAvailable
	return waitUntilObjectAvailable(t, options, availability, retries, sleepBetweenRetries)
}

func WaitUntilControllerConfigAvailable(t testing.TestingT, options *k8s.KubectlOptions, name string, retries int, sleepBetweenRetries time.Duration) (*unstructured.Unstructured, error) {
	resource := schema.GroupVersionResource{Group: "pkg.crossplane.io", Version: "v1alpha1", Resource: "controllerconfigs"}
	return waitUntilObjectAvailable(t, options, defaultObjectAvailability(name, resource), retries, sleepBetweenRetries)
}

func WaitUntilProviderConfigAvailable(t testing.TestingT, options *k8s.KubectlOptions, name string, retries int, sleepBetweenRetries time.Duration) (*unstructured.Unstructured, error) {
	resource := schema.GroupVersionResource{Group: "aws.crossplane.io", Version: "v1beta1", Resource: "providerconfigs"}
	return waitUntilObjectAvailable(t, options, defaultObjectAvailability(name, resource), retries, sleepBetweenRetries)
}

func CreateS3Bucket(t testing.TestingT, options *k8s.KubectlOptions, name string, templateFile string) (*unstructured.Unstructured, error) {
	logger.Logf(t, "Creating S3 bucket %s", name)
	bucketObject, err := readObjectFromFile(templateFile)
	if err != nil {
		return nil, err
	}
	bucketObject.SetName(name)
	resource := schema.GroupVersionResource{Group: "s3.aws.crossplane.io", Version: "v1beta1", Resource: "buckets"}
	return createObject(t, options, bucketObject, resource)
}

func DeleteS3Bucket(t testing.TestingT, options *k8s.KubectlOptions, name string) error {
	logger.Logf(t, "Deleting S3 bucket %s", name)
	resource := schema.GroupVersionResource{Group: "s3.aws.crossplane.io", Version: "v1beta1", Resource: "buckets"}
	return deleteObject(t, options, name, resource)
}

type objectAvailability struct {
	name        string
	namespace   string
	resource    schema.GroupVersionResource
	isAvailable isObjectAvailable
	objectError NewObjectError
}

func defaultObjectAvailability(name string, resource schema.GroupVersionResource) objectAvailability {
	return objectAvailability{
		name:        name,
		namespace:   "",
		resource:    resource,
		isAvailable: isObjectNotNil,
		objectError: DefaultObjectError,
	}
}

type isObjectAvailable func(*unstructured.Unstructured) bool

func waitUntilObjectAvailable(
	t testing.TestingT,
	options *k8s.KubectlOptions,
	availability objectAvailability,
	retries int,
	sleepBetweenRetries time.Duration,
) (*unstructured.Unstructured, error) {
	statusMsg := fmt.Sprintf("Wait for %s %s to be provisioned.", availability.resource.Resource, availability.name)
	var object *unstructured.Unstructured
	message, err := retry.DoWithRetryE(t, statusMsg, retries, sleepBetweenRetries, func() (string, error) {
		provider, err := getObject(t, options, availability.name, availability.namespace, availability.resource)
		if err != nil {
			return "", err
		}
		if !availability.isAvailable(provider) {
			return "", availability.objectError(provider)
		}
		object = provider
		return fmt.Sprintf("%s %s is now available", availability.resource.Resource, availability.name), nil
	},
	)
	if err != nil {
		logger.Logf(t, "Timed out waiting for %s %s to be provisioned: %s", availability.resource.Resource,
			availability.name, err)
		return nil, err
	}
	logger.Logf(t, message)
	return object, nil
}

func getObject(t testing.TestingT, options *k8s.KubectlOptions, name string, namespace string, resource schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	dynamicClient, err := GetDynamicKubernetesClientFromOptionsE(t, options)
	if err != nil {
		return nil, err
	}
	return dynamicClient.Resource(resource).Namespace(namespace).Get(context.Background(), name, metaV1.GetOptions{})
}

func createObject(t testing.TestingT, options *k8s.KubectlOptions, object *unstructured.Unstructured, resource schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	dynamicClient, err := GetDynamicKubernetesClientFromOptionsE(t, options)
	if err != nil {
		return nil, err
	}
	return dynamicClient.Resource(resource).Create(context.Background(), object, metaV1.CreateOptions{})
}

func deleteObject(t testing.TestingT, options *k8s.KubectlOptions, name string, resource schema.GroupVersionResource) error {
	dynamicClient, err := GetDynamicKubernetesClientFromOptionsE(t, options)
	if err != nil {
		return err
	}
	return dynamicClient.Resource(resource).Delete(context.Background(), name, metaV1.DeleteOptions{})
}

func readObjectFromFile(templateFile string) (*unstructured.Unstructured, error) {
	var object unstructured.Unstructured
	bytes, err := os.ReadFile(templateFile)
	if err != nil {
		return nil, err
	}
	err = k8sYaml.Unmarshal(bytes, &object)
	if err != nil {
		return nil, err
	}
	if object.Object == nil {
		return nil, fmt.Errorf("failed to read object from file %s", templateFile)
	}
	return &object, nil
}

func isProviderAvailable(provider *unstructured.Unstructured) bool {
	status := getProviderStatus(provider)
	return status["Healthy"] == "True" && status["Installed"] == "True"
}

func isObjectNotNil(config *unstructured.Unstructured) bool {
	return config != nil && config.Object != nil
}

func getProviderStatus(provider *unstructured.Unstructured) map[string]string {
	conditions, found, err := unstructured.NestedSlice(provider.Object, "status", "conditions")
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

func GetStringValue(object map[string]interface{}, fieldStrings ...string) string {
	value, found, err := unstructured.NestedString(object, fieldStrings...)
	if !found || err != nil {
		return ""
	}
	return value
}
