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
	})
}
