package test

import (
	"fmt"
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

	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)

	_, err := k8s.WaitUntilK8SGatewayAvailable(t, kubectlOptions, fmt.Sprintf("%s-public", namespaceName), 50, 6*time.Second)
	require.NoError(t, err, "oracle-gateway not available error")

	switch envName {
	case "biz":
		// pri disables the private gateway, see test/oracle_pri.yaml
		_, err = k8s.WaitUntilK8SGatewayAvailable(t, kubectlOptions, fmt.Sprintf("%s-private", namespaceName), 50, 6*time.Second)
		require.NoError(t, err, "oracle-gateway not available error")
	}
}
