package test

import (
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/require"
)

func TestK8sGatewayApiCrdsOracleDev(t *testing.T) {
	testK8sGatewayApiCrds(t, "oracle", "dev")
}

func testK8sGatewayApiCrds(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, _ := k8s.CheckKubectlConnection(t, cloudName, envName)

	// Istio 1.30 needs the v1 versions of these served, TLSRoute and ReferenceGrant
	// included - with older CRDs istiod ignores those two silently rather than erroring.
	err := k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "gateway.networking.k8s.io/v1",
		[]string{"gatewayclasses", "gateways", "httproutes", "tlsroutes", "referencegrants"}, 60, 2*time.Second)
	require.NoError(t, err, "Gateway API CRDs error")
}
