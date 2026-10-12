package test

import (
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/require"
)

// TODO(oracle): re-enable once a shared Oracle OKE test cluster exists in CI.
// There is no OKE cluster behind these tests yet, so they only ever fail in
// CheckKubectlConnection.
/*
func TestK8sOracleGatewayBiz(t *testing.T) {
	testK8sOracleGateway(t, "oracle", "biz")
}

func TestK8sOracleGatewayPri(t *testing.T) {
	testK8sOracleGateway(t, "oracle", "pri")
}
*/

func testK8sOracleGateway(t *testing.T, cloudName string, envName string) {
	t.Parallel()

	kubectlOptions, _ := k8s.CheckKubectlConnection(t, cloudName, envName)

	_, err := k8s.WaitUntilK8SGatewayAvailable(t, kubectlOptions, "external", 50, 6*time.Second)
	require.NoError(t, err, "oracle-gateway not available error")

	switch envName {
	case "biz":
		// pri disables the internal gateway, see test/oracle_pri.yaml
		_, err = k8s.WaitUntilK8SGatewayAvailable(t, kubectlOptions, "internal", 50, 6*time.Second)
		require.NoError(t, err, "oracle-gateway not available error")
	}
}
