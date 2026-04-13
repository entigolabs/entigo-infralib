package test

import (
	"fmt"
	"testing"
	"time"
	"strings"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const domain = "infralib.entigo.io"

func TestK8sAwsAlbGatewayApiBiz(t *testing.T) {
	testK8sAwsAlbGatewayApi(t, "aws", "biz")
}

func TestK8sAwsAlbGatewayApiPri(t *testing.T) {
	testK8sAwsAlbGatewayApi(t, "aws", "pri")
}

func testK8sAwsAlbGatewayApi(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)

	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, fmt.Sprintf("%s-aws-load-balancer-controller", namespaceName), 10, 6*time.Second)

	uniqueId := strings.ToLower(random.UniqueId())
	resourceName := fmt.Sprintf("%s-%s", namespaceName, uniqueId)
	gatewayClassName := resourceName
	hostname := fmt.Sprintf("%s.%s", uniqueId, domain)

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

	gatewayClass, err := k8s.ReadObjectFromFile(t, "./templates/gatewayclass.yaml")
	require.NoError(t, err)
	gatewayClass.SetName(gatewayClassName)
	createdGatewayClass, err := k8s.CreateK8SGatewayClass(t, kubectlOptions, gatewayClass)
	require.NoError(t, err, "Creating GatewayClass error")
	assert.NotNil(t, createdGatewayClass)

	gateway, err := k8s.ReadObjectFromFile(t, "./templates/gateway.yaml")
	require.NoError(t, err)
	gateway.SetName(resourceName)
	err = unstructured.SetNestedField(gateway.Object, gatewayClassName, "spec", "gatewayClassName")
	require.NoError(t, err)
	createdGateway, err := k8s.CreateK8SGateway(t, kubectlOptions, gateway)
	require.NoError(t, err, "Creating Gateway error")
	assert.NotNil(t, createdGateway)

	createdGateway, err = k8s.WaitUntilK8SGatewayAvailable(t, kubectlOptions, resourceName, 60, 5*time.Second)
	if err != nil {
		_ = k8s.DeleteK8SGateway(t, kubectlOptions, resourceName)
	}
	require.NoError(t, err, "Gateway availability error")

	gatewayAddress := k8s.GetK8SGatewayAddress(createdGateway)
	require.NotEmpty(t, gatewayAddress, "Gateway address is empty")

	httpRoute, err := k8s.ReadObjectFromFile(t, "./templates/httproute.yaml")
	require.NoError(t, err)
	httpRoute.SetName(resourceName)
	err = k8s.SetNestedSliceString(httpRoute.Object, 0, "name", resourceName, "spec", "parentRefs")
	require.NoError(t, err, "Setting HTTPRoute parentRef name error")
	err = unstructured.SetNestedStringSlice(httpRoute.Object, []string{hostname}, "spec", "hostnames")
	require.NoError(t, err, "Setting HTTPRoute hostnames error")
	rules, found, err := unstructured.NestedSlice(httpRoute.Object, "spec", "rules")
	require.NoError(t, err)
	require.True(t, found, "HTTPRoute spec.rules not found")
	backendRefs, ok := rules[0].(map[string]interface{})["backendRefs"].([]interface{})
	require.True(t, ok, "HTTPRoute spec.rules[0].backendRefs not found")
	backendRefs[0].(map[string]interface{})["name"] = resourceName
	rules[0].(map[string]interface{})["backendRefs"] = backendRefs
	err = unstructured.SetNestedSlice(httpRoute.Object, rules, "spec", "rules")
	require.NoError(t, err, "Setting HTTPRoute backendRef name error")
	createdHTTPRoute, err := k8s.CreateK8SHTTPRoute(t, kubectlOptions, httpRoute)
	require.NoError(t, err, "Creating HTTPRoute error")
	assert.NotNil(t, createdHTTPRoute)

	_, err = k8s.WaitUntilK8SHTTPRouteAvailable(t, kubectlOptions, resourceName, 60, 5*time.Second)
	if err != nil {
		_ = k8s.DeleteK8SHTTPRoute(t, kubectlOptions, resourceName)
		_ = k8s.DeleteK8SGateway(t, kubectlOptions, resourceName)
	}
	require.NoError(t, err, "HTTPRoute availability error")

	targetURL := fmt.Sprintf("http://%s/", hostname)
	err = k8s.WaitUntilHostnameAvailableWithAddress(t, kubectlOptions, 60, 5*time.Second, gatewayAddress, namespaceName, targetURL, "200")
	if err != nil {
		_ = k8s.DeleteK8SHTTPRoute(t, kubectlOptions, resourceName)
		_ = k8s.DeleteK8SGateway(t, kubectlOptions, resourceName)
	}
	require.NoError(t, err, "HTTPRoute HTTP test error")

	err = k8s.DeleteK8SHTTPRoute(t, kubectlOptions, resourceName)
	require.NoError(t, err, "Deleting HTTPRoute error")
	err = k8s.WaitUntilK8SHTTPRouteDeleted(t, kubectlOptions, resourceName, 40, 2*time.Second)
	require.NoError(t, err, "HTTPRoute didn't get deleted")

	err = k8s.DeleteK8SGateway(t, kubectlOptions, resourceName)
	require.NoError(t, err, "Deleting Gateway error")
	err = k8s.WaitUntilK8SGatewayDeleted(t, kubectlOptions, resourceName, 40, 2*time.Second)
	require.NoError(t, err, "Gateway didn't get deleted")

	err = k8s.DeleteK8SGatewayClass(t, kubectlOptions, gatewayClassName)
	require.NoError(t, err, "Deleting GatewayClass error")

	err = k8s.DeleteK8SService(t, kubectlOptions, resourceName)
	require.NoError(t, err, "Deleting Service error")
	err = k8s.DeleteK8SDeployment(t, kubectlOptions, resourceName)
	require.NoError(t, err, "Deleting Deployment error")
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
