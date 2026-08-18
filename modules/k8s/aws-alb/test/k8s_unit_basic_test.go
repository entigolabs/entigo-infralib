package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// GatewayClass and Gateways are provisioned by the module itself,
// the test only verifies their health and attaches HTTPRoutes.
const gatewayClassName = "alb"

func TestK8sAwsAlbGatewayApiBiz(t *testing.T) {
	testK8sAwsAlbGatewayApi(t, "aws", "biz", []string{"external", "internal", "service"})
}

func TestK8sAwsAlbGatewayApiPri(t *testing.T) {
	// service gateway is disabled in pri
	testK8sAwsAlbGatewayApi(t, "aws", "pri", []string{"external", "internal"})
}

func testK8sAwsAlbGatewayApi(t *testing.T, cloudName string, envName string, gatewayNames []string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)
	_, _, hostName, _ := k8s.GetGatewayConfig(t, cloudName, envName, "default")

	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, fmt.Sprintf("%s-aws-load-balancer-controller", namespaceName), 10, 6*time.Second)

	// Health check of the module-provisioned GatewayClass: Accepted must be True
	acceptedStatus, err := terrak8s.RunKubectlAndGetOutputE(t, kubectlOptions,
		"get", "gatewayclass", gatewayClassName,
		"-o", `jsonpath={.status.conditions[?(@.type=="Accepted")].status}`)
	require.NoError(t, err, "Getting GatewayClass %s error", gatewayClassName)
	require.Equal(t, "True", acceptedStatus, "GatewayClass %s is not Accepted", gatewayClassName)

	// One shared backend workload for all gateway checks
	resourceName := fmt.Sprintf("%s-%s", namespaceName, strings.ToLower(random.UniqueId()))

	deployment, err := k8s.ReadObjectFromFile(t, "./templates/deployment.yaml")
	require.NoError(t, err)
	deployment.SetName(resourceName)
	_, err = k8s.CreateK8SDeployment(t, kubectlOptions, deployment)
	require.NoError(t, err, "Creating Deployment error")

	service, err := k8s.ReadObjectFromFile(t, "./templates/service.yaml")
	require.NoError(t, err)
	service.SetName(resourceName)
	_, err = k8s.CreateK8SService(t, kubectlOptions, service)
	require.NoError(t, err, "Creating Service error")

	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, resourceName, 20, 6*time.Second)

	// Every enabled gateway gets its own HTTPRoute with a unique hostname
	for _, gatewayName := range gatewayNames {
		testGatewayHTTPRoute(t, kubectlOptions, namespaceName, gatewayName, hostName, resourceName)
	}

	err = k8s.DeleteK8SService(t, kubectlOptions, resourceName)
	require.NoError(t, err, "Deleting Service error")
	err = k8s.DeleteK8SDeployment(t, kubectlOptions, resourceName)
	require.NoError(t, err, "Deleting Deployment error")
}

func testGatewayHTTPRoute(t *testing.T, kubectlOptions *terrak8s.KubectlOptions, namespaceName string, gatewayName string, hostName string, backendName string) {
	// Health check of the module-provisioned Gateway, the helper waits for Programmed
	existingGateway, err := k8s.WaitUntilK8SGatewayAvailable(t, kubectlOptions, gatewayName, 60, 5*time.Second)
	require.NoError(t, err, "Gateway %s availability error", gatewayName)

	// The Gateway must reference the expected GatewayClass
	className, found, err := unstructured.NestedString(existingGateway.Object, "spec", "gatewayClassName")
	require.NoError(t, err)
	require.True(t, found, "Gateway %s spec.gatewayClassName not found", gatewayName)
	require.Equal(t, gatewayClassName, className, "Gateway %s does not use GatewayClass %s", gatewayName, gatewayClassName)

	gatewayAddress := k8s.GetK8SGatewayAddress(existingGateway)
	require.NotEmpty(t, gatewayAddress, "Gateway %s address is empty", gatewayName)

	// Randomized hostname, same mechanism as the ingress test
	hostname := fmt.Sprintf("%s-%s", strings.ToLower(random.UniqueId()), hostName)
	routeName := fmt.Sprintf("%s-%s", backendName, gatewayName)

	httpRoute, err := k8s.ReadObjectFromFile(t, "./templates/httproute.yaml")
	require.NoError(t, err)
	httpRoute.SetName(routeName)
	err = k8s.SetNestedSliceString(httpRoute.Object, 0, "name", gatewayName, "spec", "parentRefs")
	require.NoError(t, err, "Setting HTTPRoute parentRef name error")
	err = k8s.SetNestedSliceString(httpRoute.Object, 0, "namespace", namespaceName, "spec", "parentRefs")
	require.NoError(t, err, "Setting HTTPRoute parentRef namespace error")
	err = unstructured.SetNestedStringSlice(httpRoute.Object, []string{hostname}, "spec", "hostnames")
	require.NoError(t, err, "Setting HTTPRoute hostnames error")
	rules, found, err := unstructured.NestedSlice(httpRoute.Object, "spec", "rules")
	require.NoError(t, err)
	require.True(t, found, "HTTPRoute spec.rules not found")
	backendRefs, ok := rules[0].(map[string]interface{})["backendRefs"].([]interface{})
	require.True(t, ok, "HTTPRoute spec.rules[0].backendRefs not found")
	backendRefs[0].(map[string]interface{})["name"] = backendName
	rules[0].(map[string]interface{})["backendRefs"] = backendRefs
	err = unstructured.SetNestedSlice(httpRoute.Object, rules, "spec", "rules")
	require.NoError(t, err, "Setting HTTPRoute backendRef name error")
	createdHTTPRoute, err := k8s.CreateK8SHTTPRoute(t, kubectlOptions, httpRoute)
	require.NoError(t, err, "Creating HTTPRoute for gateway %s error", gatewayName)
	assert.NotNil(t, createdHTTPRoute)

	_, err = k8s.WaitUntilK8SHTTPRouteAvailable(t, kubectlOptions, routeName, 60, 5*time.Second)
	if err != nil {
		_ = k8s.DeleteK8SHTTPRoute(t, kubectlOptions, routeName)
	}
	require.NoError(t, err, "HTTPRoute availability error for gateway %s", gatewayName)

	targetURL := fmt.Sprintf("http://%s/", hostname)
	err = k8s.WaitUntilHostnameAvailableWithAddress(t, kubectlOptions, 60, 5*time.Second, gatewayAddress, namespaceName, targetURL, "200")
	if err != nil {
		_ = k8s.DeleteK8SHTTPRoute(t, kubectlOptions, routeName)
	}
	require.NoError(t, err, "HTTPRoute HTTP test error for gateway %s", gatewayName)

	err = k8s.DeleteK8SHTTPRoute(t, kubectlOptions, routeName)
	require.NoError(t, err, "Deleting HTTPRoute error for gateway %s", gatewayName)
	err = k8s.WaitUntilK8SHTTPRouteDeleted(t, kubectlOptions, routeName, 40, 2*time.Second)
	require.NoError(t, err, "HTTPRoute didn't get deleted for gateway %s", gatewayName)
}

func TestK8sAwsAlbBiz(t *testing.T) {
	testK8sAwsAlb(t, "aws", "biz")
}

func TestK8sAwsAlbPri(t *testing.T) {
	testK8sAwsAlb(t, "aws", "pri")
}

func testK8sAwsAlb(t *testing.T, cloudName string, envName string) {

	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)
	_, _, hostName, _ := k8s.GetGatewayConfig(t, cloudName, envName, "default")

	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, fmt.Sprintf("%s-aws-load-balancer-controller", namespaceName), 10, 6*time.Second)
	terrak8s.WaitUntilServiceAvailable(t, kubectlOptions, "aws-load-balancer-webhook-service", 60, 1*time.Second)
	time.Sleep(5 * time.Second)

	ingress, err := k8s.ReadObjectFromFile(t, "./templates/ingress.yaml")
	require.NoError(t, err)
	ingress.SetName(fmt.Sprintf("%s-%s", namespaceName, strings.ToLower(random.UniqueId())))
	ingressClass := "alb"
	err = unstructured.SetNestedField(ingress.Object, ingressClass, "spec", "ingressClassName")
	require.NoError(t, err, "Setting ingressClassName error")
	annotations := ingress.GetAnnotations()
	annotations["alb.ingress.kubernetes.io/group.name"] = "aws-load-balancer"
	ingress.SetAnnotations(annotations)
	host := fmt.Sprintf("%s-%s", strings.ToLower(random.UniqueId()), hostName)
	err = k8s.SetNestedSliceString(ingress.Object, 0, "host", host, "spec", "rules")
	require.NoError(t, err, "Setting host error")
	createdIngress, err := k8s.CreateK8SIngress(t, kubectlOptions, ingress)
	require.NoError(t, err, "Creating ingress error")
	assert.NotNil(t, createdIngress, "Ingress is nil")

	_, err = k8s.WaitUntilK8SIngressAvailable(t, kubectlOptions, createdIngress.GetName(), 40, 2*time.Second)
	if err != nil {
		_ = k8s.DeleteK8SIngress(t, kubectlOptions, ingress.GetName()) // Try to delete ingress
	}
	require.NoError(t, err, "Ingress availability error")

	err = k8s.DeleteK8SIngress(t, kubectlOptions, ingress.GetName())
	require.NoError(t, err, "Deleting ingress error")

	err = k8s.WaitUntilK8SIngressDeleted(t, kubectlOptions, ingress.GetName(), 40, 2*time.Second)
	require.NoError(t, err, "Ingress didn't get deleted")
}
