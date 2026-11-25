package test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
	"github.com/entigolabs/entigo-infralib-test/retry"
)

// fullnameOverride pins the Deployment and the Service name to "kube-state-metrics".
const name = "kube-state-metrics"

// One metric of every kind of object the dashboards and alerts read, all of which exist in
// every test cluster. A kind that kube-state-metrics may not list (RBAC) or cannot reach
// (NetworkPolicy to the API) has no series at all.
var families = []string{
	"kube_node_info",
	"kube_namespace_status_phase",
	"kube_pod_info",
	"kube_pod_status_phase",
	"kube_pod_container_resource_requests",
	"kube_deployment_status_replicas_available",
	"kube_daemonset_status_number_ready",
	"kube_statefulset_replicas",
	"kube_service_info",
}

func TestKubeStateMetrics(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, name, 20, 6*time.Second), "kube-state-metrics deployment error")

		// A running pod is not yet one that reads the cluster: without the API it serves an
		// empty page. Read what the scrapers read.
		addr, stop := k8s.PortForwardService(t, c, name, 8080)
		defer stop()

		// Its own Deployment, as the API reports it now.
		self := fmt.Sprintf(`kube_deployment_status_replicas_available{namespace=%q,deployment=%q} 1`, c.Namespace, name)

		_, err := retry.DoWithRetryE(t, "kube-state-metrics serves the cluster's state", 20, 6*time.Second, func() (string, error) {
			metrics, err := get(fmt.Sprintf("http://%s/metrics", addr))
			if err != nil {
				return "", err
			}
			return "", missing(metrics, self)
		})
		require.NoError(t, err)
	})
}

// missing names what a /metrics page lacks: a metric family, the allowlisted node labels, or
// the line of kube-state-metrics' own Deployment.
func missing(metrics, self string) error {
	var absent []string
	nodeLabels := false
	seen := map[string]bool{}
	for _, line := range strings.Split(metrics, "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}
		family, _, _ := strings.Cut(line, "{")
		seen[family] = true
		// metricLabelsAllowlist: every node has this label, and it is only there when allowed.
		if family == "kube_node_labels" && strings.Contains(line, `label_kubernetes_io_arch="`) {
			nodeLabels = true
		}
		if line == self {
			self = ""
		}
	}
	for _, family := range families {
		if !seen[family] {
			absent = append(absent, family)
		}
	}
	if !nodeLabels {
		absent = append(absent, "kube_node_labels with label_kubernetes_io_arch")
	}
	if self != "" {
		absent = append(absent, self)
	}
	if len(absent) > 0 {
		return fmt.Errorf("kube-state-metrics does not report: %s", strings.Join(absent, ", "))
	}
	return nil
}

func get(url string) (string, error) {
	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s returned %d", url, resp.StatusCode)
	}
	return string(body), nil
}
