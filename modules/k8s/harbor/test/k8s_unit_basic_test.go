package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

func TestHarbor(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		for _, pod := range []string{"database-0", "redis-0", "trivy-0"} {
			name := fmt.Sprintf("%s-%s", c.Namespace, pod)
			require.NoError(t, k8s.WaitUntilPodAvailableE(t, c, name, 20, 6*time.Second), "%s pod error", name)
		}
		for _, deployment := range []string{"core", "portal"} {
			name := fmt.Sprintf("%s-%s", c.Namespace, deployment)
			require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, name, 20, 6*time.Second), "%s deployment error", name)
		}
		for _, deployment := range []string{"registry", "jobservice"} {
			name := fmt.Sprintf("%s-%s", c.Namespace, deployment)
			require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, name, 30, 10*time.Second), "%s deployment error", name)
		}

		// The harbor HTTPRoute is the only one in the namespace.
		k8s.WaitUntilRouteAnswers(t, c, "", "/api/v2.0/ping", "200", routeRetries(e), 6*time.Second)
	})
}

// Google and Azure gateways take long to program a new route.
func routeRetries(e *env.Environment) int {
	if e.Cloud == env.CloudGoogle || e.Cloud == env.CloudAzure {
		return 400
	}
	return 100
}
