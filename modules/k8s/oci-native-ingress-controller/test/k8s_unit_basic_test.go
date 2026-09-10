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
)

const controllerNamespace = "native-ingress-controller-system"

func TestK8sOciNativeIngressControllerDev(t *testing.T) {
	testK8sOciNativeIngressController(t, "oracle", "dev")
}

func testK8sOciNativeIngressController(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)

	// The controller runs in a fixed namespace (native-ingress-controller-system, set by
	// values.yaml's fullnameOverride/deploymentNamespace), independent of wherever ArgoCD
	// places this Application - not the module's own app namespace.
	controllerOptions := terrak8s.NewKubectlOptions(kubectlOptions.ContextName, kubectlOptions.ConfigPath, controllerNamespace)
	terrak8s.WaitUntilDeploymentAvailable(t, controllerOptions, "oci-native-ingress-controller", 20, 6*time.Second)

	deployment, err := k8s.ReadObjectFromFile(t, "./templates/deployment.yaml")
	require.NoError(t, err)
	_, err = k8s.CreateK8SDeployment(t, kubectlOptions, deployment)
	require.NoError(t, err, "Creating Deployment error")

	service, err := k8s.ReadObjectFromFile(t, "./templates/service.yaml")
	require.NoError(t, err)
	_, err = k8s.CreateK8SService(t, kubectlOptions, service)
	require.NoError(t, err, "Creating Service error")

	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, deployment.GetName(), 20, 6*time.Second)

	ingress, err := k8s.ReadObjectFromFile(t, "./templates/ingress.yaml")
	require.NoError(t, err)
	ingressName := fmt.Sprintf("%s-%s", namespaceName, strings.ToLower(random.UniqueId()))
	ingress.SetName(ingressName)
	err = k8s.SetNestedSliceString(ingress.Object, 0, "host", fmt.Sprintf("%s.test", ingressName), "spec", "rules")
	require.NoError(t, err, "Setting host error")

	createdIngress, err := k8s.CreateK8SIngress(t, kubectlOptions, ingress)
	require.NoError(t, err, "Creating Ingress error")
	assert.NotNil(t, createdIngress, "Ingress is nil")

	_, err = k8s.WaitUntilK8SIngressAvailable(t, kubectlOptions, ingressName, 60, 5*time.Second)
	if err != nil {
		_ = k8s.DeleteK8SIngress(t, kubectlOptions, ingressName)
	}
	require.NoError(t, err, "Ingress availability error - expected an OCI Load Balancer IP in status.loadBalancer.ingress[0].ip")

	err = k8s.DeleteK8SIngress(t, kubectlOptions, ingressName)
	require.NoError(t, err, "Deleting Ingress error")
	err = k8s.WaitUntilK8SIngressDeleted(t, kubectlOptions, ingressName, 40, 2*time.Second)
	require.NoError(t, err, "Ingress didn't get deleted")

	err = k8s.DeleteK8SService(t, kubectlOptions, service.GetName())
	require.NoError(t, err, "Deleting Service error")
	err = k8s.DeleteK8SDeployment(t, kubectlOptions, deployment.GetName())
	require.NoError(t, err, "Deleting Deployment error")
}
