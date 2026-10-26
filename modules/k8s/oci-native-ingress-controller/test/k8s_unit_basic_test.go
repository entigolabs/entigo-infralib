package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
	"github.com/entigolabs/entigo-infralib-test/random"
)

const controllerNamespace = "native-ingress-controller-system"

func TestOciNativeIngressController(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		// The controller runs in a fixed namespace (native-ingress-controller-system, set by
		// values.yaml's fullnameOverride/deploymentNamespace), independent of wherever ArgoCD
		// places this Application - not the module's own app namespace.
		controller := c.WithNamespace(controllerNamespace)
		k8s.WaitUntilDeploymentAvailable(t, controller, "oci-native-ingress-controller", 20, 6*time.Second)

		deployment, err := k8s.ReadObjectFromFile(t, "./templates/deployment.yaml")
		require.NoError(t, err)
		_, err = k8s.CreateK8SDeployment(t, c, deployment)
		require.NoError(t, err, "Creating Deployment error")

		service, err := k8s.ReadObjectFromFile(t, "./templates/service.yaml")
		require.NoError(t, err)
		_, err = k8s.CreateK8SService(t, c, service)
		require.NoError(t, err, "Creating Service error")

		k8s.WaitUntilDeploymentAvailable(t, c, deployment.GetName(), 20, 6*time.Second)

		ingress, err := k8s.ReadObjectFromFile(t, "./templates/ingress.yaml")
		require.NoError(t, err)
		ingressName := fmt.Sprintf("%s-%s", c.Namespace, strings.ToLower(random.UniqueId()))
		ingress.SetName(ingressName)
		err = k8s.SetNestedSliceString(ingress.Object, 0, "host", fmt.Sprintf("%s.test", ingressName), "spec", "rules")
		require.NoError(t, err, "Setting host error")

		createdIngress, err := k8s.CreateK8SIngress(t, c, ingress)
		require.NoError(t, err, "Creating Ingress error")
		assert.NotNil(t, createdIngress, "Ingress is nil")

		_, err = k8s.WaitUntilK8SIngressAvailable(t, c, ingressName, 60, 5*time.Second)
		if err != nil {
			_ = k8s.DeleteK8SIngress(t, c, ingressName)
		}
		require.NoError(t, err, "Ingress availability error - expected an OCI Load Balancer IP in status.loadBalancer.ingress[0].ip")

		err = k8s.DeleteK8SIngress(t, c, ingressName)
		require.NoError(t, err, "Deleting Ingress error")
		err = k8s.WaitUntilK8SIngressDeleted(t, c, ingressName, 40, 2*time.Second)
		require.NoError(t, err, "Ingress didn't get deleted")

		err = k8s.DeleteK8SService(t, c, service.GetName())
		require.NoError(t, err, "Deleting Service error")
		err = k8s.DeleteK8SDeployment(t, c, deployment.GetName())
		require.NoError(t, err, "Deleting Deployment error")
	})
}
