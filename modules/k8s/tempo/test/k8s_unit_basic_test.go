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

	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/aws"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/google"
	"github.com/entigolabs/entigo-infralib-test/k8s"
	"github.com/entigolabs/entigo-infralib-test/retry"
)

// fullnameOverride in values.yaml pins every chart resource name to "tempo".
const tempoName = "tempo"

func TestTempo(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		// The upstream chart runs tempo as a single-replica StatefulSet, not a Deployment.
		err := k8s.WaitUntilPodAvailableE(t, c, fmt.Sprintf("%s-0", tempoName), 20, 6*time.Second)
		require.NoError(t, err, "%s pod error", c.Namespace)

		// Tempo's usage reporter (chart default reportingEnabled: true) writes a cluster
		// seed file to the trace backend on startup, which proves the pod can write to
		// the bucket with its cloud identity.
		switch e.Cloud {
		case env.CloudAWS:
			bucket := fmt.Sprintf("%s-%s-%s-%s", e.Prefix, c.Namespace, aws.AccountID(t, e), e.MustRegion(t))
			err = aws.WaitUntilBucketFileAvailable(t, e.MustRegion(t), bucket, "tempo_cluster_seed.json", 20, 6*time.Second)
			require.NoError(t, err, "tempo_cluster_seed.json not found in AWS bucket")
		case env.CloudGoogle:
			err = google.WaitUntilBucketFileAvailable(t, fmt.Sprintf("%s-%s-traces-%s", e.Prefix, c.Namespace, e.MustProject(t)), "tempo_cluster_seed.json", 20, 6*time.Second)
			require.NoError(t, err, "tempo_cluster_seed.json not found in Google bucket")
		}

		otlpAddr, stopOtlp := k8s.PortForwardService(t, c, tempoName, 4318)
		defer stopOtlp()

		apiAddr, stopAPI := k8s.PortForwardService(t, c, tempoName, 3200)
		defer stopAPI()

		apiURL := fmt.Sprintf("http://%s", apiAddr)
		otlpURL := fmt.Sprintf("http://%s/v1/traces", otlpAddr)

		_, err = retry.DoWithRetryE(t, "Wait for tempo /ready", 20, 6*time.Second, func() (string, error) {
			return httpDo(http.MethodGet, apiURL+"/ready", nil)
		})
		require.NoError(t, err, "tempo never became ready")

		traceID := randomHex(t, 16)
		serviceName := fmt.Sprintf("infralib-test-%s-%s", e.Cloud, e.Prefix)
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
	})
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
