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

func WaitUntilDeploymentRuntimeConfigAvailable(t testing.TestingT, options *k8s.KubectlOptions, name string, retries int, sleepBetweenRetries time.Duration) (*unstructured.Unstructured, error) {
	resource := schema.GroupVersionResource{Group: "pkg.crossplane.io", Version: "v1beta1", Resource: "deploymentruntimeconfigs"}
	return waitUntilObjectAvailable(t, options, defaultObjectAvailability(name, resource), retries, sleepBetweenRetries)
}

func WaitUntilProviderConfigAvailable(t testing.TestingT, options *k8s.KubectlOptions, resource schema.GroupVersionResource, name string, retries int, sleepBetweenRetries time.Duration) (*unstructured.Unstructured, error) {
	return waitUntilObjectAvailable(t, options, defaultObjectAvailability(name, resource), retries, sleepBetweenRetries)
}

func WaitUntilK8SBucketAvailable(t testing.TestingT, options *k8s.KubectlOptions, name string, retries int, sleepBetweenRetries time.Duration) (*unstructured.Unstructured, error) {
	resource := schema.GroupVersionResource{Group: "s3.aws.crossplane.io", Version: "v1beta1", Resource: "buckets"}
	availability := defaultObjectAvailability(name, resource)
	availability.isAvailable = isCrossplaneObjectAvailable
	availability.objectError = NewCrossplaneObjectNotAvailable
	return waitUntilObjectAvailable(t, options, availability, retries, sleepBetweenRetries)
}

func WaitUntilK8SBucketDeleted(t testing.TestingT, options *k8s.KubectlOptions, name string, retries int, sleepBetweenRetries time.Duration) error {
	resource := schema.GroupVersionResource{Group: "s3.aws.crossplane.io", Version: "v1beta1", Resource: "buckets"}
	namespacedObject := defaultNamespacedObject(name, resource)
	return waitUntilObjectDeleted(t, options, namespacedObject, retries, sleepBetweenRetries)
}

func CreateK8SBucket(t testing.TestingT, options *k8s.KubectlOptions, name string, templateFile string) (*unstructured.Unstructured, error) {
	logger.Logf(t, "Creating S3 bucket %s", name)
	bucketObject, err := ReadObjectFromFile(t, templateFile)
	if err != nil {
		return nil, err
	}
	bucketObject.SetName(name)
	resource := schema.GroupVersionResource{Group: "s3.aws.crossplane.io", Version: "v1beta1", Resource: "buckets"}
	return createObject(t, options, bucketObject, "", resource)
}

func DeleteK8SBucket(t testing.TestingT, options *k8s.KubectlOptions, name string) error {
	logger.Logf(t, "Deleting S3 bucket %s", name)
	resource := schema.GroupVersionResource{Group: "s3.aws.crossplane.io", Version: "v1beta1", Resource: "buckets"}
	return deleteObject(t, options, name, "", resource)
}

func WaitUntilK8SObjectAvailable(t testing.TestingT, options *k8s.KubectlOptions, name string, retries int, sleepBetweenRetries time.Duration) (*unstructured.Unstructured, error) {
	resource := schema.GroupVersionResource{Group: "kubernetes.crossplane.io", Version: "v1alpha2", Resource: "objects"}
	availability := defaultObjectAvailability(name, resource)
	availability.isAvailable = isCrossplaneObjectAvailable
	availability.objectError = NewCrossplaneObjectNotAvailable
	return waitUntilObjectAvailable(t, options, availability, retries, sleepBetweenRetries)
}

func WaitUntilK8SObjectDeleted(t testing.TestingT, options *k8s.KubectlOptions, name string, retries int, sleepBetweenRetries time.Duration) error {
	resource := schema.GroupVersionResource{Group: "kubernetes.crossplane.io", Version: "v1alpha2", Resource: "objects"}
	namespacedObject := defaultNamespacedObject(name, resource)
	return waitUntilObjectDeleted(t, options, namespacedObject, retries, sleepBetweenRetries)
}

func CreateK8SObject(t testing.TestingT, options *k8s.KubectlOptions, name string, templateFile string) (*unstructured.Unstructured, error) {
	logger.Logf(t, "Creating k8s provider object %s", name)
	object, err := ReadObjectFromFile(t, templateFile)
	if err != nil {
		return nil, err
	}
	object.SetName(name)
	err = unstructured.SetNestedField(object.Object, name, "spec", "forProvider", "manifest", "metadata", "name")
	if err != nil {
		return nil, err
	}
	err = unstructured.SetNestedField(object.Object, options.Namespace, "spec", "forProvider", "manifest", "metadata", "namespace")
	if err != nil {
		return nil, err
	}
	resource := schema.GroupVersionResource{Group: "kubernetes.crossplane.io", Version: "v1alpha2", Resource: "objects"}
	return createObject(t, options, object, "", resource)
}

func DeleteK8SObject(t testing.TestingT, options *k8s.KubectlOptions, name string) error {
	logger.Logf(t, "Deleting k8s provider object %s", name)
	resource := schema.GroupVersionResource{Group: "kubernetes.crossplane.io", Version: "v1alpha2", Resource: "objects"}
	return deleteObject(t, options, name, "", resource)
}

func WaitUntilK8SIngressAvailable(t testing.TestingT, options *k8s.KubectlOptions, name string, retries int, sleepBetweenRetries time.Duration) (*unstructured.Unstructured, error) {
	resource := schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}
	availability := defaultObjectAvailability(name, resource)
	availability.namespacedObject.namespace = options.Namespace
	availability.isAvailable = isIngressAvailable
	availability.objectError = NewIngressNotAvailable
	return waitUntilObjectAvailable(t, options, availability, retries, sleepBetweenRetries)
}

func WaitUntilK8SIngressDeleted(t testing.TestingT, options *k8s.KubectlOptions, name string, retries int, sleepBetweenRetries time.Duration) error {
	resource := schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}
	namespacedObject := defaultNamespacedObject(name, resource)
	namespacedObject.namespace = options.Namespace
	return waitUntilObjectDeleted(t, options, namespacedObject, retries, sleepBetweenRetries)
}

func CreateK8SIngress(t testing.TestingT, options *k8s.KubectlOptions, ingressObject *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	logger.Logf(t, "Creating Ingress %s", ingressObject.GetName())
	resource := schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}
	return createObject(t, options, ingressObject, options.Namespace, resource)
}

func DeleteK8SIngress(t testing.TestingT, options *k8s.KubectlOptions, name string) error {
	logger.Logf(t, "Deleting Ingress %s", name)
	resource := schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}
	return deleteObject(t, options, name, options.Namespace, resource)
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

func createObject(t testing.TestingT, options *k8s.KubectlOptions, object *unstructured.Unstructured, namespace string, resource schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	dynamicClient, err := GetDynamicKubernetesClientFromOptionsE(t, options)
	if err != nil {
		return nil, err
	}
	return dynamicClient.Resource(resource).Namespace(namespace).Create(context.Background(), object, metaV1.CreateOptions{})
}

func deleteObject(t testing.TestingT, options *k8s.KubectlOptions, name string, namespace string, resource schema.GroupVersionResource) error {
	dynamicClient, err := GetDynamicKubernetesClientFromOptionsE(t, options)
	if err != nil {
		return err
	}
	return dynamicClient.Resource(resource).Namespace(namespace).Delete(context.Background(), name, metaV1.DeleteOptions{})
}

func ReadObjectFromFile(t testing.TestingT, templateFile string) (*unstructured.Unstructured, error) {
	logger.Log(t, fmt.Sprintf("Reading k8s object from file %s", templateFile))
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

func isCrossplaneObjectAvailable(object *unstructured.Unstructured) bool {
	status := getStatusMap(object)
	return status["Ready"] == "True" && status["Synced"] == "True"
}

func isIngressAvailable(ingress *unstructured.Unstructured) bool {
	ingresses, found, err := unstructured.NestedSlice(ingress.Object, "status", "loadBalancer", "ingress")
	if !found || err != nil || len(ingresses) == 0 {
		return false
	}
	ingressMap := ingresses[0].(map[string]interface{})
	return ingressMap["hostname"] != ""
}

func getStatusMap(object *unstructured.Unstructured) map[string]string {
	conditions, found, err := unstructured.NestedSlice(object.Object, "status", "conditions")
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

func SetNestedSliceString(object map[string]interface{}, index int, label string, value string, fieldStrings ...string) error {
	rules, found, err := unstructured.NestedSlice(object, fieldStrings...)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("field %s not found", fieldStrings)
	}
	if (index + 1) > len(rules) {
		return fmt.Errorf("index %d out of range", index)
	}
	rules[index].(map[string]interface{})[label] = value
	return unstructured.SetNestedSlice(object, rules, fieldStrings...)
}
