package test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Lines of a real /metrics page of alloy v1.22.0.
const page = `# TYPE loki_write_sent_entries_total counter
alloy_component_controller_running_components{controller_id="",controller_path="/",health_type="healthy"} 5
loki_write_sent_entries_total{component_id="loki.write.default",component_path="/",host="loki-gateway.loki"} 12
prometheus_remote_storage_queue_highest_sent_timestamp_seconds{component_id="prometheus.remote_write.default",component_path="/",remote_name="3827c5",url="http://mimir-gateway.mimir/api/v1/push"} 1.795685941e+09`

func TestUndelivered(t *testing.T) {
	require.NoError(t, undelivered(page, true))
	require.ErrorContains(t, undelivered(strings.ReplaceAll(page, `"} 12`, `"} 0`), true), "no log line accepted")
	require.ErrorContains(t, undelivered(strings.ReplaceAll(page, "1.795685941e+09", "0"), true), "no sample accepted")
	// Metrics off: no endpoint, nothing to hold against it.
	require.NoError(t, undelivered(strings.ReplaceAll(page, "prometheus_remote_storage", "x"), true))
	unhealthy := page + "\n" + `alloy_component_controller_running_components{controller_id="",controller_path="/",health_type="unhealthy"} 1`
	require.ErrorContains(t, undelivered(unhealthy, true), "not healthy")
	// Without stores only the components count.
	require.NoError(t, undelivered(strings.ReplaceAll(page, "1.795685941e+09", "0"), false))
	require.ErrorContains(t, undelivered(unhealthy, false), "not healthy")
}
