package test

import (
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestK8sOracleGatewayOracleDev(t *testing.T) {
	testK8sOracleGateway(t, "oracle", "dev")
}

func testK8sOracleGateway(t *testing.T, cloudName string, envName string) {
	t.Parallel()

	releaseName := "oracle-gateway"

	kubectlOptions, _ := k8s.CheckKubectlConnection(t, cloudName, envName)
	kubectlOptions.Namespace = releaseName

	// Waits for an address on the Gateway, which Istio only publishes once it has
	// provisioned the Service and the OCI load balancer has been assigned an IP.
	gateway, err := k8s.WaitUntilK8SGatewayAvailable(t, kubectlOptions, releaseName, 90, 10*time.Second)
	require.NoError(t, err, "Gateway available error")
	assert.NotNil(t, gateway, "Gateway is nil")

	// The address is what external-dns publishes for every HTTPRoute attached here, so an
	// empty one means DNS would silently point nowhere.
	assert.NotEmpty(t, k8s.GetK8SGatewayAddress(gateway), "Gateway address is empty")

	// Istio names the generated data plane <gateway>-<gatewayclass>.
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, releaseName+"-istio", 30, 10*time.Second)
	require.NoError(t, err, "Istio gateway deployment error")
}
