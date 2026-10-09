package test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestK8sCertManagerAzureBiz(t *testing.T) {
	testK8sCertManager(t, "azure", "biz")
}

func TestK8sCertManagerAzurePri(t *testing.T) {
	testK8sCertManager(t, "azure", "pri")
}

func testK8sCertManager(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)

	for _, deployment := range []string{namespaceName, fmt.Sprintf("%s-cainjector", namespaceName), fmt.Sprintf("%s-webhook", namespaceName)} {
		err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, deployment, 60, 10*time.Second)
		require.NoError(t, err, "%s deployment error", deployment)
	}

	clusterIssuerGVR := schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "clusterissuers"}
	_, err := waitUntilConditions(t, kubectlOptions, clusterIssuerGVR, "", "letsencrypt", []string{"Ready"}, 60, 10*time.Second)
	require.NoError(t, err, "ClusterIssuer error")

	client, err := k8s.GetDynamicKubernetesClientFromOptionsE(t, kubectlOptions)
	require.NoError(t, err, "Kubernetes client error")
	_, _, hostName, _ := k8s.GetGatewayConfig(t, cloudName, envName, "external")
	domain := strings.SplitN(hostName, ".", 2)[1]
	randomName := strings.ToLower(random.UniqueId())
	certificateName := fmt.Sprintf("%s-%s", namespaceName, randomName)
	secretName := fmt.Sprintf("%s-tls", certificateName)
	certificate := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "cert-manager.io/v1",
		"kind":       "Certificate",
		"metadata": map[string]interface{}{
			"name":      certificateName,
			"namespace": namespaceName,
		},
		"spec": map[string]interface{}{
			"secretName": secretName,
			"dnsNames":   []interface{}{fmt.Sprintf("%s.%s", randomName, domain)},
			"issuerRef": map[string]interface{}{
				"kind": "ClusterIssuer",
				"name": "letsencrypt",
			},
		},
	}}
	certificateGVR := schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "certificates"}
	_, err = client.Resource(certificateGVR).Namespace(namespaceName).Create(context.Background(), certificate, metaV1.CreateOptions{})
	require.NoError(t, err, "Creating Certificate error")
	defer func() {
		_ = client.Resource(certificateGVR).Namespace(namespaceName).Delete(context.Background(), certificateName, metaV1.DeleteOptions{})
		secretGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
		_ = client.Resource(secretGVR).Namespace(namespaceName).Delete(context.Background(), secretName, metaV1.DeleteOptions{})
	}()

	_, err = waitUntilConditions(t, kubectlOptions, certificateGVR, namespaceName, certificateName, []string{"Ready"}, 90, 10*time.Second)
	require.NoError(t, err, "Certificate %s error", certificateName)
}

func waitUntilConditions(t *testing.T, options *terrak8s.KubectlOptions, resource schema.GroupVersionResource, namespace string, name string, conditions []string, retries int, sleepBetweenRetries time.Duration) (*unstructured.Unstructured, error) {
	client, err := k8s.GetDynamicKubernetesClientFromOptionsE(t, options)
	if err != nil {
		return nil, err
	}
	var object *unstructured.Unstructured
	_, err = retry.DoWithRetryE(t, fmt.Sprintf("Wait for %s %s", resource.Resource, name), retries, sleepBetweenRetries, func() (string, error) {
		object, err = client.Resource(resource).Namespace(namespace).Get(context.Background(), name, metaV1.GetOptions{})
		if err != nil {
			return "", err
		}
		status := map[string]string{}
		items, _, _ := unstructured.NestedSlice(object.Object, "status", "conditions")
		for _, item := range items {
			if condition, ok := item.(map[string]interface{}); ok {
				status[fmt.Sprint(condition["type"])] = fmt.Sprint(condition["status"])
			}
		}
		for _, condition := range conditions {
			if status[condition] != "True" {
				return "", fmt.Errorf("%s %s condition %s is %q", resource.Resource, name, condition, status[condition])
			}
		}
		return fmt.Sprintf("%s %s is available", resource.Resource, name), nil
	})
	return object, err
}
