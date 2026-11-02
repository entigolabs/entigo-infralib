package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

func TestArgocd(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		require.NoError(t, k8s.WaitUntilResourcesAvailable(t, c, "argoproj.io/v1alpha1", []string{"applications"}, 60, 1*time.Second), "Argocd no Applications CRD")
		require.NoError(t, k8s.WaitUntilResourcesAvailable(t, c, "argoproj.io/v1alpha1", []string{"applicationsets"}, 60, 1*time.Second), "Argocd no Applicationsets CRD")

		for _, component := range []string{"server", "repo-server", "notifications-controller", "applicationset-controller"} {
			require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, c.Namespace+"-"+component, 20, 6*time.Second), "argocd-%s deployment error", component)
		}
		// Dex is only deployed in the biz environments; pri uses direct OIDC.
		if e.Prefix != "pri" {
			require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, c.Namespace+"-dex-server", 20, 6*time.Second), "argocd-dex-server deployment error")
		}

		// The server's HTTPRoute is the only one in the namespace; its hostname,
		// listener and gateway address all come from the cluster.
		k8s.WaitUntilRouteReachable(t, c, "", routeRetries(e), 6*time.Second)
	})
}

// Google and Azure gateways take long to program a new route.
func routeRetries(e *env.Environment) int {
	if e.Cloud == env.CloudGoogle || e.Cloud == env.CloudAzure {
		return 400
	}
	return 100
}
