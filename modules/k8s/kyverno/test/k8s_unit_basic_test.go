package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

func TestKyverno(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, "kyverno-admission-controller", 10, 6*time.Second), "kyverno-admission-controller deployment error")
		// require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, "kyverno-background-controller", 10, 6*time.Second), "kyverno-background-controller deployment error")
		// require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, "kyverno-cleanup-controller", 10, 6*time.Second), "kyverno-cleanup-controller deployment error")
		// require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, "kyverno-reports-controller", 10, 6*time.Second), "kyverno-reports-controller deployment error")
	})
}
