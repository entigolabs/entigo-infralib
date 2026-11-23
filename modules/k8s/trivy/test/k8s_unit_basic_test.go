package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

func TestTrivy(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		// The trivy-operator subchart names its deployment <release>-trivy-operator.
		require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, c.Namespace+"-trivy-operator", 20, 6*time.Second), "trivy-operator deployment error")

		// The built-in trivy server only answers its readiness probe once it has downloaded
		// the vulnerability database, on AWS through the ECR pull through cache.
		require.NoError(t, k8s.WaitUntilPodAvailableE(t, c, "trivy-server-0", 50, 6*time.Second), "trivy-server pod error")
	})
}
