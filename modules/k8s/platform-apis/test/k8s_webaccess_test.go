package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
)

// ── Orchestrator ──────────────────────────────────────────────────────────────

func testWebAccess(t *testing.T, ctx context.Context, cluster, argocd *terrak8s.KubectlOptions) {
	waNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, WebAccessNamespaceName)
	defer cleanupWebAccess(t, cluster, argocd)

	if ctx.Err() != nil {
		return
	}
	applyFile(t, cluster, "./templates/webaccess_test_application.yaml")
	syncWithRetry(t, argocd, WebAccessApplicationName)
	if ctx.Err() != nil {
		return
	}

	t.Run("WebAccess", func(t *testing.T) { testWebAccessResource(t, waNs) })
}

// ── WebAccess ─────────────────────────────────────────────────────────────────

func testWebAccessResource(t *testing.T, waNs *terrak8s.KubectlOptions) {
	t.Helper()

	// Create: wait for composite to be Synced+Ready
	waitSyncedAndReady(t, waNs, WebAccessKind, WebAccessName, 30, 10*time.Second)
	if t.Failed() {
		return
	}

	// Sub-resources: Istio resources must be created
	t.Run("SubResources", func(t *testing.T) {
		t.Run("VirtualService", func(t *testing.T) {
			t.Parallel()
			waitResourceExists(t, waNs, WebAccessVirtualSvcKind, WebAccessName, 15, 10*time.Second)
		})
		t.Run("ServiceEntry", func(t *testing.T) {
			t.Parallel()
			waitWebAccessServiceEntryExists(t, waNs)
		})
		t.Run("DestinationRule", func(t *testing.T) {
			t.Parallel()
			waitWebAccessDestinationRuleExists(t, waNs)
		})
	})
	if t.Failed() {
		return
	}

	// Read: verify VirtualService fields
	require.Equal(t, "test-webapp-service",
		getField(t, waNs, WebAccessVirtualSvcKind, WebAccessName, ".spec.hosts[0]"))
	require.Equal(t, "test-webaccess.example.com",
		getField(t, waNs, WebAccessVirtualSvcKind, WebAccessName, ".spec.hosts[1]"))
	require.Equal(t, "/",
		getField(t, waNs, WebAccessVirtualSvcKind, WebAccessName, ".spec.http[0].match[0].uri.prefix"))
	require.Equal(t, "test-webapp-service",
		getField(t, waNs, WebAccessVirtualSvcKind, WebAccessName, ".spec.http[0].route[0].destination.host"))
	require.Equal(t, "80",
		getField(t, waNs, WebAccessVirtualSvcKind, WebAccessName, ".spec.http[0].route[0].destination.port.number"))
}

func waitWebAccessServiceEntryExists(t *testing.T, waNs *terrak8s.KubectlOptions) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, "ServiceEntry for test-webaccess", 15, 10*time.Second,
		func() (string, error) {
			out, err := terrak8s.RunKubectlAndGetOutputE(t, waNs, "get", WebAccessServiceEntryKind,
				"-l", fmt.Sprintf("crossplane.io/composite=%s", WebAccessName),
				"-o", "jsonpath={.items[0].metadata.name}")
			if err != nil || out == "" {
				return "", fmt.Errorf("no ServiceEntry for %s yet", WebAccessName)
			}
			return out, nil
		})
	require.NoError(t, err)
}

func waitWebAccessDestinationRuleExists(t *testing.T, waNs *terrak8s.KubectlOptions) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, "DestinationRule for test-webaccess", 15, 10*time.Second,
		func() (string, error) {
			out, err := terrak8s.RunKubectlAndGetOutputE(t, waNs, "get", WebAccessDestRuleKind,
				"-l", fmt.Sprintf("crossplane.io/composite=%s", WebAccessName),
				"-o", "jsonpath={.items[0].metadata.name}")
			if err != nil || out == "" {
				return "", fmt.Errorf("no DestinationRule for %s yet", WebAccessName)
			}
			return out, nil
		})
	require.NoError(t, err)
}

// ── Cleanup ───────────────────────────────────────────────────────────────────

func cleanupWebAccess(t *testing.T, cluster, argocd *terrak8s.KubectlOptions) {
	if t.Failed() {
		return
	}
	waNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, WebAccessNamespaceName)

	cleanupDeleteAndWait(t, waNs, WebAccessKind, WebAccessName, 30)

	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocd, "delete", "application", WebAccessApplicationName, "--ignore-not-found")
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, cluster, "delete", "namespace", WebAccessNamespaceName, "--ignore-not-found", "--wait=true")
}
