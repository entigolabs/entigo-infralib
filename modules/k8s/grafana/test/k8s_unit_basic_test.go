package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

func TestGrafana(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, "grafana", 20, 6*time.Second), "grafana deployment error")

		// The grafana HTTPRoute is the only one in the namespace.
		k8s.WaitUntilRouteAnswers(t, c, "", "/login", "200", routeRetries(e), 6*time.Second)
	})
}

// Google and Azure gateways take long to program a new route.
func routeRetries(e *env.Environment) int {
	if e.Cloud == env.CloudGoogle || e.Cloud == env.CloudAzure {
		return 400
	}
	return 100
}
