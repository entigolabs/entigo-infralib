package k8s

import (
	"context"
	"errors"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	kubernetesErrors "k8s.io/apimachinery/pkg/api/errors"
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

func WaitUntilK8SBucketAvailable(t testing.TestingT, options *k8s.KubectlOptions, name string, retries int, sleepBetweenRetries time.Duration) (*unstructured.Unstructured, error) {
	resource := schema.GroupVersionResource{Group: "s3.aws.crossplane.io", Version: "v1beta1", Resource: "buckets"}
	availability := defaultObjectAvailability(name, resource)
	availability.isAvailable = isBucketAvailable
	availability.objectError = NewBucketNotAvailable
	return waitUntilObjectAvailable(t, options, availability, retries, sleepBetweenRetries)
}

func WaitUntilK8SBucketDeleted(t testing.TestingT, options *k8s.KubectlOptions, name string, retries int, sleepBetweenRetries time.Duration) error {
	resource := schema.GroupVersionResource{Group: "s3.aws.crossplane.io", Version: "v1beta1", Resource: "buckets"}
	namespacedObject := defaultNamespacedObject(name, resource)
	return waitUntilObjectDeleted(t, options, namespacedObject, retries, sleepBetweenRetries)
}

func CreateK8SBucket(t testing.TestingT, options *k8s.KubectlOptions, name string, templateFile string) (*unstructured.Unstructured, error) {
	logger.Logf(t, "Creating S3 bucket %s", name)
	bucketObject, err := readObjectFromFile(templateFile)
	if err != nil {
		return nil, err
	}
	bucketObject.SetName(name)
	resource := schema.GroupVersionResource{Group: "s3.aws.crossplane.io", Version: "v1beta1", Resource: "buckets"}
	return createObject(t, options, bucketObject, resource)
}

func DeleteK8SBucket(t testing.TestingT, options *k8s.KubectlOptions, name string) error {
	logger.Logf(t, "Deleting S3 bucket %s", name)
	resource := schema.GroupVersionResource{Group: "s3.aws.crossplane.io", Version: "v1beta1", Resource: "buckets"}
	return deleteObject(t, options, name, resource)
}

type objectAvailability struct {
	namespacedObject namespacedObject
	isAvailable      isObjectAvailable
	objectError      NewObjectError
}

func defaultObjectAvailability(name string, resource schema.GroupVersionResource) objectAvailability {
	return objectAvailability{
		namespacedObject: defaultNamespacedObject(name, resource),
		isAvailable:      isObjectNotNil,
		objectError:      DefaultObjectError,
	}
}

type namespacedObject struct {
	name      string
	namespace string
	resource  schema.GroupVersionResource
}

func defaultNamespacedObject(name string, resource schema.GroupVersionResource) namespacedObject {
	return namespacedObject{
		name:      name,
		namespace: "",
		resource:  resource,
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
	namespacedObject := availability.namespacedObject
	statusMsg := fmt.Sprintf("Wait for %s %s to be provisioned.", namespacedObject.resource.Resource, namespacedObject.name)
	var object *unstructured.Unstructured
	message, err := retry.DoWithRetryE(t, statusMsg, retries, sleepBetweenRetries, func() (string, error) {
		var err error
		object, err = getObject(t, options, namespacedObject.name, namespacedObject.namespace, namespacedObject.resource)
		if err != nil {
			return "", err
		}
		if !availability.isAvailable(object) {
			return "", availability.objectError(object)
		}
		return fmt.Sprintf("%s %s is now available", namespacedObject.resource.Resource, namespacedObject.name), nil
	},
	)
	if err != nil {
		logger.Logf(t, "Timed out waiting for %s %s to be provisioned: %s", namespacedObject.resource.Resource,
			namespacedObject.name, err)
		return nil, err
	}
	logger.Logf(t, message)
	return object, nil
}

func waitUntilObjectDeleted(
	t testing.TestingT,
	options *k8s.KubectlOptions,
	namespacedObject namespacedObject,
	retries int,
	sleepBetweenRetries time.Duration,
) error {
	statusMsg := fmt.Sprintf("Wait for %s %s to be deleted.", namespacedObject.resource.Resource, namespacedObject.name)
	message, err := retry.DoWithRetryE(t, statusMsg, retries, sleepBetweenRetries, func() (string, error) {
		_, err := getObject(t, options, namespacedObject.name, namespacedObject.namespace, namespacedObject.resource)
		if err == nil {
			return "", fmt.Errorf("%s %s still exists", namespacedObject.resource.Resource, namespacedObject.name)
		}
		var statusError *kubernetesErrors.StatusError
		if errors.As(err, &statusError) && statusError.Status().Code == 404 {
			return fmt.Sprintf("%s %s is now deleted", namespacedObject.resource.Resource, namespacedObject.name), nil
		}
		return "", err
	},
	)
	if err != nil {
		logger.Logf(t, "Timed out waiting for %s %s to be deleted: %s", namespacedObject.resource.Resource,
			namespacedObject.name, err)
		return err
	}
	logger.Logf(t, message)
	return nil
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
	status := getStatusMap(provider)
	return status["Healthy"] == "True" && status["Installed"] == "True"
}

func isObjectNotNil(config *unstructured.Unstructured) bool {
	return config != nil && config.Object != nil
}

func isBucketAvailable(bucket *unstructured.Unstructured) bool {
	status := getStatusMap(bucket)
	return status["Ready"] == "True" && status["Synced"] == "True"
}

func getStatusMap(provider *unstructured.Unstructured) map[string]string {
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
