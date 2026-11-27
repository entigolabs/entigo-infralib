package test

import (
	"io"
	"net/http"
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
		if e.Prefix != "biz" {
			require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, c.Namespace+"-kube-state-metrics", 20, 6*time.Second), "%s-kube-state-metrics deployment error", c.Namespace)
		}
		require.NoError(t, k8s.WaitUntilPodAvailableE(t, c, c.Namespace+"-alertmanager-0", 20, 6*time.Second), "%s-alertmanager-0 pod error", c.Namespace)

		// The server's HTTPRoute (prometheus.server.route from the subchart) is the only one in the namespace.
		k8s.WaitUntilRouteAnswers(t, c, "", "/query", "200", routeRetries(e), 6*time.Second)

		// The remote write receiver is on for the cluster and not reachable through the route.
		k8s.WaitUntilRouteAnswers(t, c, "", "/api/v1/write", "404", routeRetries(e), 6*time.Second)
		addr, stop := k8s.PortForwardService(t, c, c.Namespace+"-server", 80)
		defer stop()
		// An empty write is refused as malformed (400) by a receiver that is on; one that is
		// off answers 404.
		resp, err := http.Post("http://"+addr+"/api/v1/write", "application/x-protobuf", nil)
		require.NoError(t, err)
		resp.Body.Close()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode, "the remote write receiver is not on")

		// biz runs Prometheus as a store that alloy writes to (test/<cloud>_biz.yaml).
		if e.Prefix == "biz" {
			resp, err = http.Get("http://" + addr + "/api/v1/targets?state=active")
			require.NoError(t, err)
			defer resp.Body.Close()
			targets, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Contains(t, string(targets), `"job":"prometheus"`, "Prometheus should still scrape itself")
			require.NotContains(t, string(targets), `"job":"kubernetes-`, "a store must not scrape the cluster")
		}
	})
}

// Google and Azure gateways take long to program a new route.
func routeRetries(e *env.Environment) int {
	if e.Cloud == env.CloudGoogle || e.Cloud == env.CloudAzure {
		return 400
	}
	return 100
}
