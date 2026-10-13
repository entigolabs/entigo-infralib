package test

import (
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/require"
)

func TestK8sOracleGatewayBiz(t *testing.T) {
	testK8sOracleGateway(t, "oracle", "biz")
}

func TestK8sOracleGatewayPri(t *testing.T) {
	testK8sOracleGateway(t, "oracle", "pri")
}

func testK8sOracleGateway(t *testing.T, cloudName string, envName string) {
	t.Parallel()

	kubectlOptions, _ := k8s.CheckKubectlConnection(t, cloudName, envName)

	// Only the internal gateway is deployed, see test/oracle_biz.yaml.
	_, err := k8s.WaitUntilK8SGatewayAvailable(t, kubectlOptions, "internal", 50, 6*time.Second)
	require.NoError(t, err, "oracle-gateway not available error")
}
