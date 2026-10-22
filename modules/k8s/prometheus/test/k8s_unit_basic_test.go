package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

func TestPrometheus(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, c.Namespace+"-server", 20, 6*time.Second), "%s-server deployment error", c.Namespace)
		require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, c.Namespace+"-kube-state-metrics", 20, 6*time.Second), "%s-kube-state-metrics deployment error", c.Namespace)
		require.NoError(t, k8s.WaitUntilPodAvailableE(t, c, c.Namespace+"-alertmanager-0", 20, 6*time.Second), "%s-alertmanager-0 pod error", c.Namespace)

		// The server's HTTPRoute (prometheus.server.route from the subchart) is the only one in the namespace.
		k8s.WaitUntilRouteAnswers(t, c, "", "/query", "200", routeRetries(e), 6*time.Second)
	})
}

// Google and Azure gateways take long to program a new route.
func routeRetries(e *env.Environment) int {
	if e.Cloud == env.CloudGoogle || e.Cloud == env.CloudAzure {
		return 400
	}
	return 100
}
