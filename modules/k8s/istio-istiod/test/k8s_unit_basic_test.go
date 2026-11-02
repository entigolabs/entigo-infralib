package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

func TestIstioIstiod(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		// The application is named istio-system in every environment, so the
		// client's namespace is istio-system.
		c := k8s.Connect(t, e)

		output, err := k8s.RunKubectlAndGetOutputE(t, c, "auth", "can-i", "get", "pods")
		require.NoError(t, err, "Unable to connect to context %s cluster %s", c.Context, err)
		require.Equal(t, "yes", output)

		require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, "istiod", 10, 6*time.Second), "istiod deployment error")

		_, err = k8s.RunKubectlAndGetOutputE(t, c, "get", "envoyfilter", "remove-envoy-headers")
		require.NoError(t, err, "remove-envoy-headers EnvoyFilter missing: %s", err)
	})
}
