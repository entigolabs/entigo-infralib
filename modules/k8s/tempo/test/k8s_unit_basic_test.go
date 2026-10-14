package test

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
)

// fullnameOverride in values.yaml pins every chart resource name to "tempo".
const tempoName = "tempo"

func TestK8sTempoAWSBiz(t *testing.T) {
	testK8sTempo(t, "aws", "biz")
}

func TestK8sTempoAWSPri(t *testing.T) {
	testK8sTempo(t, "aws", "pri")
}

func TestK8sTempoGoogleBiz(t *testing.T) {
	testK8sTempo(t, "google", "biz")
}

func testK8sTempo(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)

	// The upstream chart runs tempo as a single-replica StatefulSet, not a Deployment.
	err := terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, fmt.Sprintf("%s-0", tempoName), 20, 6*time.Second)
	require.NoError(t, err, fmt.Sprintf("%s pod error", namespaceName))

	// Tempo's usage reporter (chart default reportingEnabled: true) writes a cluster
	// seed file to the trace backend on startup, which proves the pod can write to
	// the bucket with its cloud identity.
	switch cloudName {
	case "aws":
		err = aws.WaitUntilBucketFileAvailable(t, fmt.Sprintf("%s-%s-877483565445-eu-north-1", envName, namespaceName), "tempo_cluster_seed.json", 20, 6*time.Second)
		require.NoError(t, err, "tempo_cluster_seed.json not found in AWS bucket")
	case "google":
		err = google.WaitUntilBucketFileAvailable(t, fmt.Sprintf("%s-%s-traces-entigo-infralib2", envName, namespaceName), "tempo_cluster_seed.json", 20, 6*time.Second)
		require.NoError(t, err, "tempo_cluster_seed.json not found in Google bucket")
	}

	otlpTunnel := terrak8s.NewTunnel(kubectlOptions, terrak8s.ResourceTypeService, tempoName, 0, 4318)
	defer otlpTunnel.Close()
	otlpTunnel.ForwardPort(t)

	apiTunnel := terrak8s.NewTunnel(kubectlOptions, terrak8s.ResourceTypeService, tempoName, 0, 3200)
	defer apiTunnel.Close()
	apiTunnel.ForwardPort(t)

	apiURL := fmt.Sprintf("http://%s", apiTunnel.Endpoint())
	otlpURL := fmt.Sprintf("http://%s/v1/traces", otlpTunnel.Endpoint())

	_, err = retry.DoWithRetryE(t, "Wait for tempo /ready", 20, 6*time.Second, func() (string, error) {
		return httpDo(http.MethodGet, apiURL+"/ready", nil)
	})
	require.NoError(t, err, "tempo never became ready")

	traceID := randomHex(t, 16)
	serviceName := fmt.Sprintf("infralib-test-%s-%s", cloudName, envName)
	payload := otlpTracePayload(traceID, randomHex(t, 8), serviceName)

	// The ingester ring can take a moment to accept writes after /ready.
	_, err = retry.DoWithRetryE(t, "Push test span over OTLP/HTTP", 10, 6*time.Second, func() (string, error) {
		return httpDo(http.MethodPost, otlpURL, payload)
	})
	require.NoError(t, err, "failed to push test span to tempo")

	body, err := retry.DoWithRetryE(t, fmt.Sprintf("Query trace %s", traceID), 20, 6*time.Second, func() (string, error) {
		return httpDo(http.MethodGet, fmt.Sprintf("%s/api/traces/%s", apiURL, traceID), nil)
	})
	require.NoError(t, err, "pushed trace was never queryable from tempo")
	require.Contains(t, body, serviceName, "queried trace does not contain the pushed span")
}

func httpDo(method string, url string, body []byte) (string, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s %s returned %d: %s", method, url, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return string(respBody), nil
}

func randomHex(t *testing.T, n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	require.NoError(t, err)
	return hex.EncodeToString(b)
}

// OTLP/JSON encodes trace and span ids as hex and timestamps as nanosecond strings.
func otlpTracePayload(traceID string, spanID string, serviceName string) []byte {
	end := time.Now()
	start := end.Add(-time.Second)
	return []byte(fmt.Sprintf(`{
  "resourceSpans": [{
    "resource": {"attributes": [{"key": "service.name", "value": {"stringValue": %q}}]},
    "scopeSpans": [{
      "scope": {"name": "entigo-infralib-test"},
      "spans": [{
        "traceId": %q,
        "spanId": %q,
        "name": "tempo-unit-test",
        "kind": 1,
        "startTimeUnixNano": "%d",
        "endTimeUnixNano": "%d"
      }]
    }]
  }]
}`, serviceName, traceID, spanID, start.UnixNano(), end.UnixNano()))
}
