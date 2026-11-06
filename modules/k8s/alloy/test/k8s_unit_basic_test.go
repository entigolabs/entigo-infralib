package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

func TestAlloy(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		nodeAgentDaemonSetName := fmt.Sprintf("%s-node-agent", c.Namespace)

		daemonSet, err := k8s.GetDaemonSetE(t, c, nodeAgentDaemonSetName)
		require.NoError(t, err, "Daemonset %s error", nodeAgentDaemonSetName)
		assert.NotEmpty(t, daemonSet, "Node-agent daemonset was not returned")

		// cluster-metrics/receiver are opt-in (they need a Mimir / otlp-or-faro destination), off
		// by default. The biz envs enable them (see test/*_biz.yaml); pri runs logs-only, so
		// neither is present there.
		if e.Prefix == "biz" {
			require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, c.Namespace+"-cluster-metrics", 20, 6*time.Second), "cluster-metrics deployment error")
			require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, c.Namespace+"-receiver", 20, 6*time.Second), "receiver deployment error")
		}
	})
}
