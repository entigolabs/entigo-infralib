package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

func TestClusterAutoscaler(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		// The upstream chart names the deployment <release>-<cloudProvider>-cluster-autoscaler
		// (verified via helm template for both values files).
		provider := map[string]string{env.CloudAWS: "aws", env.CloudOracle: "oci"}[e.Cloud]
		deploymentName := fmt.Sprintf("%s-%s-cluster-autoscaler", c.Namespace, provider)
		err := k8s.WaitUntilDeploymentAvailableE(t, c, deploymentName, 50, 6*time.Second)
		require.NoError(t, err, "cluster-autoscaler deployment %s error: %s", c.Namespace, err)
	})
}
