package test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const self = `kube_deployment_status_replicas_available{namespace="ksm",deployment="kube-state-metrics"} 1`

func page() string {
	lines := []string{
		"# HELP kube_node_info Information about a cluster node.",
		`kube_node_labels{node="a",label_kubernetes_io_arch="amd64"} 1`,
		self,
	}
	for _, family := range families {
		lines = append(lines, family+`{namespace="x"} 1`)
	}
	return strings.Join(lines, "\n")
}

func TestMissing(t *testing.T) {
	require.NoError(t, missing(page(), self))

	// A pod that cannot reach the API serves only its own telemetry.
	require.ErrorContains(t, missing("", self), "kube_node_info")

	require.ErrorContains(t, missing(strings.ReplaceAll(page(), "kube_statefulset_replicas", "x"), self), "kube_statefulset_replicas")
	require.ErrorContains(t, missing(strings.ReplaceAll(page(), "label_kubernetes_io_arch", "x"), self), "label_kubernetes_io_arch")
	require.ErrorContains(t, missing(strings.ReplaceAll(page(), `"kube-state-metrics"} 1`, `"kube-state-metrics"} 0`), self), "replicas_available")
}
