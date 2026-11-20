package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

func TestKubeStateMetrics(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		// fullnameOverride pins the Deployment name to "kube-state-metrics".
		require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, "kube-state-metrics", 20, 6*time.Second), "kube-state-metrics deployment error")
	})
}
