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

func TestK8sAzureGatewayBiz(t *testing.T) {
	testK8sAzureGateway(t, "azure", "biz")
}

func TestK8sAzureGatewayPri(t *testing.T) {
	testK8sAzureGateway(t, "azure", "pri")
}

func testK8sAzureGateway(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)
	gatewayName := fmt.Sprintf("%s-external", namespaceName)

	albOptions := terrak8s.NewKubectlOptions(kubectlOptions.ContextName, "", "azure-alb-system")
	err := terrak8s.WaitUntilDeploymentAvailableE(t, albOptions, "alb-controller", 60, 10*time.Second)
	require.NoError(t, err, "alb-controller deployment error")

	applicationLoadBalancerGVR := schema.GroupVersionResource{Group: "alb.networking.azure.io", Version: "v1", Resource: "applicationloadbalancers"}
	_, err = waitUntilConditions(t, kubectlOptions, applicationLoadBalancerGVR, namespaceName, namespaceName, []string{"Accepted", "Deployment"}, 60, 10*time.Second)
	require.NoError(t, err, "ApplicationLoadBalancer error")

	certificateGVR := schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "certificates"}
	_, err = waitUntilConditions(t, kubectlOptions, certificateGVR, namespaceName, fmt.Sprintf("%s-tls", gatewayName), []string{"Ready"}, 90, 10*time.Second)
	require.NoError(t, err, "Certificate error")

	gateway, err := k8s.WaitUntilK8SGatewayAvailable(t, kubectlOptions, gatewayName, 120, 10*time.Second)
	require.NoError(t, err, "Gateway %s error", gatewayName)
	require.NotEmpty(t, k8s.GetK8SGatewayAddress(gateway), "Gateway %s has no address", gatewayName)

	frontendTLSPolicyGVR := schema.GroupVersionResource{Group: "alb.networking.azure.io", Version: "v1", Resource: "frontendtlspolicies"}
	_, err = waitUntilConditions(t, kubectlOptions, frontendTLSPolicyGVR, namespaceName, gatewayName, []string{"Accepted", "ResolvedRefs"}, 30, 10*time.Second)
	require.NoError(t, err, "FrontendTLSPolicy error")

	_, err = k8s.WaitUntilK8SHTTPRouteAvailable(t, kubectlOptions, fmt.Sprintf("%s-redirect", gatewayName), 60, 10*time.Second)
	require.NoError(t, err, "redirect HTTPRoute error")

	targetURL := fmt.Sprintf("http://%s.%s-net-dns.azure.infralib.entigo.io", strings.ToLower(random.UniqueId()), envName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, 60, 10*time.Second, gatewayName, namespaceName, namespaceName, targetURL, "301", cloudName)
	require.NoError(t, err, "%s redirect error", targetURL)
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
