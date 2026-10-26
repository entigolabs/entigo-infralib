package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

func TestCrossplaneCore(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		k8s.WaitUntilDeploymentAvailable(t, c, "crossplane", 10, 6*time.Second)
		k8s.WaitUntilDeploymentAvailable(t, c, "crossplane-rbac-manager", 10, 6*time.Second)

		err := k8s.WaitUntilResourcesAvailable(t, c, "pkg.crossplane.io/v1", []string{"providers"}, 60, 6*time.Second)
		require.NoError(t, err, "Providers crd error")

		err = k8s.WaitUntilResourcesAvailable(t, c, "pkg.crossplane.io/v1beta1", []string{"deploymentruntimeconfigs"}, 60, 6*time.Second)
		require.NoError(t, err, "DeploymentRuntimeConfig crd error")
	})
}
