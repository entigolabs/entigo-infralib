package test

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
	"github.com/entigolabs/entigo-infralib-test/retry"
)

func TestAlloy(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		nodeAgent := fmt.Sprintf("%s-node-agent", c.Namespace)
		daemonSet, err := k8s.GetDaemonSetE(t, c, nodeAgent)
		require.NoError(t, err, "Daemonset %s error", nodeAgent)
		assert.NotEmpty(t, daemonSet, "Node-agent daemonset was not returned")

		if e.Prefix == "biz" {
			require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, c.Namespace+"-cluster-metrics", 20, 6*time.Second), "cluster-metrics deployment error")
			require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, c.Namespace+"-receiver", 20, 6*time.Second), "receiver deployment error")
		}

		// A running pod is not yet one that delivers: read the node agent's own counters, which
		// only move when Loki and the metrics stores have accepted what it sent. The Service is
		// there for this, the test inputs turn it on (node-agent.service.enabled).
		addr, stop := k8s.PortForwardService(t, c, nodeAgent, 12345)
		defer stop()
		// Oracle has no Loki, Mimir or Prometheus to deliver to, only the components are checked.
		delivers := e.Cloud != env.CloudOracle
		_, err = retry.DoWithRetryE(t, "the node agent delivers logs and metrics", 30, 10*time.Second, func() (string, error) {
			metrics, err := get(fmt.Sprintf("http://%s/metrics", addr))
			if err != nil {
				return "", err
			}
			return "", undelivered(metrics, delivers)
		})
		require.NoError(t, err)
	})
}

// undelivered names what a node agent's /metrics page shows as not working: a component that is
// not healthy and, when the agent has stores to deliver to, no log line accepted by Loki, or a
// metrics endpoint that has accepted no sample yet. Metrics are off by default (pri), then there
// is no endpoint to look at.
func undelivered(metrics string, delivers bool) error {
	var problems []string
	logs := 0.0
	for _, line := range strings.Split(metrics, "\n") {
		series, text, _ := strings.Cut(line, " ")
		value, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
		if err != nil || strings.HasPrefix(line, "#") {
			continue
		}
		switch {
		case strings.HasPrefix(series, "alloy_component_controller_running_components{") && !strings.Contains(series, `health_type="healthy"`) && value > 0:
			problems = append(problems, fmt.Sprintf("%v components are not healthy: %s", value, series))
		case strings.HasPrefix(series, "loki_write_sent_entries_total{"):
			logs += value
		// The time of the newest sample the endpoint accepted, 0 until the first one.
		case strings.HasPrefix(series, "prometheus_remote_storage_queue_highest_sent_timestamp_seconds{"):
			if value == 0 && delivers {
				problems = append(problems, "no sample accepted yet: "+series)
			}
		}
	}
	if delivers && logs == 0 {
		problems = append(problems, "no log line accepted by Loki (loki_write_sent_entries_total)")
	}
	if len(problems) > 0 {
		return fmt.Errorf("alloy node agent: %s", strings.Join(problems, "; "))
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
