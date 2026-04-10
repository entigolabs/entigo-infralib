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
	gatewayClassName := fmt.Sprintf("%s-%s", namespaceName, uniqueId)
	gatewayName := fmt.Sprintf("%s-%s", namespaceName, uniqueId)

	gatewayClass, err := k8s.ReadObjectFromFile(t, "./templates/gatewayclass.yaml")
	require.NoError(t, err)
	gatewayClass.SetName(gatewayClassName)
	createdGatewayClass, err := k8s.CreateK8SGatewayClass(t, kubectlOptions, gatewayClass)
	require.NoError(t, err, "Creating GatewayClass error")
	assert.NotNil(t, createdGatewayClass)

	gateway, err := k8s.ReadObjectFromFile(t, "./templates/gateway.yaml")
	require.NoError(t, err)
	gateway.SetName(gatewayName)
	err = unstructured.SetNestedField(gateway.Object, gatewayClassName, "spec", "gatewayClassName")
	require.NoError(t, err)
	createdGateway, err := k8s.CreateK8SGateway(t, kubectlOptions, gateway)
	require.NoError(t, err, "Creating Gateway error")
	assert.NotNil(t, createdGateway)

	_, err = k8s.WaitUntilK8SGatewayAvailable(t, kubectlOptions, gatewayName, 60, 5*time.Second)
	if err != nil {
		_ = k8s.DeleteK8SGateway(t, kubectlOptions, gatewayName)
	}
	require.NoError(t, err, "Gateway availability error")

	err = k8s.DeleteK8SGateway(t, kubectlOptions, gatewayName)
	require.NoError(t, err, "Deleting Gateway error")
	err = k8s.WaitUntilK8SGatewayDeleted(t, kubectlOptions, gatewayName, 40, 2*time.Second)
	require.NoError(t, err, "Gateway didn't get deleted")

	err = k8s.DeleteK8SGatewayClass(t, kubectlOptions, gatewayClassName)
	require.NoError(t, err, "Deleting GatewayClass error")
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
