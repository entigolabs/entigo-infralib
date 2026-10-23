package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	_ "github.com/entigolabs/entigo-infralib-test/azure"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
	"github.com/entigolabs/entigo-infralib-test/random"
	"github.com/entigolabs/entigo-infralib-test/retry"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestAzureGateway(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)
		gatewayName := "external"

		err := k8s.WaitUntilDeploymentAvailableE(t, c.WithNamespace("azure-alb-system"), "alb-controller", 60, 10*time.Second)
		require.NoError(t, err, "alb-controller deployment error")

		applicationLoadBalancerGVR := schema.GroupVersionResource{Group: "alb.networking.azure.io", Version: "v1", Resource: "applicationloadbalancer"}
		_, err = waitUntilConditions(t, c, applicationLoadBalancerGVR, c.Namespace, c.Namespace, []string{"Accepted", "Deployment"}, 60, 10*time.Second)
		require.NoError(t, err, "ApplicationLoadBalancer error")

		certificateGVR := schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "certificates"}
		_, err = waitUntilConditions(t, c, certificateGVR, c.Namespace, fmt.Sprintf("%s-tls", gatewayName), []string{"Ready"}, 90, 10*time.Second)
		require.NoError(t, err, "Certificate error")

		gateway, err := k8s.WaitUntilK8SGatewayAvailable(t, c, gatewayName, 120, 10*time.Second)
		require.NoError(t, err, "Gateway %s error", gatewayName)
		address := k8s.GetK8SGatewayAddress(gateway)
		require.NotEmpty(t, address, "Gateway %s has no address", gatewayName)

		frontendTLSPolicyGVR := schema.GroupVersionResource{Group: "alb.networking.azure.io", Version: "v1", Resource: "frontendtlspolicies"}
		_, err = waitUntilConditions(t, c, frontendTLSPolicyGVR, c.Namespace, gatewayName, []string{"Accepted", "ResolvedRefs"}, 30, 10*time.Second)
		require.NoError(t, err, "FrontendTLSPolicy error")

		_, err = k8s.WaitUntilK8SHTTPRouteAvailable(t, c, fmt.Sprintf("%s-redirect", gatewayName), 60, 10*time.Second)
		require.NoError(t, err, "redirect HTTPRoute error")

		// The http listener redirects every name of the public zone, so a random
		// hostname under the net step's domain must answer 301 through the gateway address.
		domain := tf.GetStep(t, e, "net").String(t, "dns__pub_domain")
		targetURL := fmt.Sprintf("http://%s.%s", strings.ToLower(random.UniqueId()), domain)
		err = k8s.WaitUntilHostnameAvailableWithAddress(t, c, address, targetURL, "301", 60, 10*time.Second)
		require.NoError(t, err, "%s redirect error", targetURL)

		// Private frontend on the azure/agc bring-your-own AGC, only reachable inside the VNet
		_, err = waitUntilConditions(t, c, certificateGVR, c.Namespace, "internal-tls", []string{"Ready"}, 90, 10*time.Second)
		require.NoError(t, err, "internal Certificate error")
		internalGateway, err := k8s.WaitUntilK8SGatewayAvailable(t, c, "internal", 120, 10*time.Second)
		require.NoError(t, err, "Gateway internal error")
		require.True(t, strings.HasSuffix(k8s.GetK8SGatewayAddress(internalGateway), ".privatelink.alb.azure.com"), "Gateway internal must have a private frontend address")
	})
}

// waitUntilConditions waits until every named status condition of the object is True.
func waitUntilConditions(t *testing.T, c *k8s.Client, resource schema.GroupVersionResource, namespace string, name string, conditions []string, retries int, sleepBetweenRetries time.Duration) (*unstructured.Unstructured, error) {
	var object *unstructured.Unstructured
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("Wait for %s %s", resource.Resource, name), retries, sleepBetweenRetries, func() (string, error) {
		var err error
		object, err = c.GetObjectE(resource, namespace, name)
		if err != nil {
			return "", err
		}
		status := map[string]string{}
		items, _, _ := unstructured.NestedSlice(object.Object, "status", "conditions")
		for _, item := range items {
			if condition, ok := item.(map[string]any); ok {
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
