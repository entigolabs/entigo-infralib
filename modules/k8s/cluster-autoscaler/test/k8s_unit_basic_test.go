package test

import (
	"fmt"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestK8sClusterAutoscalerAWSBiz(t *testing.T) {
	testK8sClusterAutoscaler(t, "aws", "biz")
}

func TestK8sClusterAutoscalerOracleDev(t *testing.T) {
	testK8sClusterAutoscaler(t, "oracle", "dev")
}

func testK8sClusterAutoscaler(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)

	// The upstream chart names the deployment <release>-<cloudProvider>-cluster-autoscaler
	// (verified via helm template for both values files).
	provider := map[string]string{"aws": "aws", "oracle": "oci"}[cloudName]
	deploymentName := fmt.Sprintf("%s-%s-cluster-autoscaler", namespaceName, provider)
	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, deploymentName, 50, 6*time.Second)
	require.NoError(t, err, "cluster-autoscaler deployment %s error: %s", namespaceName, err)
}
