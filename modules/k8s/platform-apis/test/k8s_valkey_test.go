package test

import (
	"context"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

// ── Orchestrator ──────────────────────────────────────────────────────────────

func testValkey(t *testing.T, ctx context.Context, cluster, argocd *terrak8s.KubectlOptions) {
	vkNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, ValkeyNamespaceName)
	defer cleanupValkey(t, cluster, argocd)

	if ctx.Err() != nil {
		return
	}
	applyFile(t, cluster, "./templates/valkey_test_application.yaml")
	syncWithRetry(t, argocd, ValkeyApplicationName)
	if ctx.Err() != nil {
		return
	}

	t.Run("instances", func(t *testing.T) {
		t.Run("CustomValkeyInstance", func(t *testing.T) { t.Parallel(); testCustomValkeyInstance(t, vkNs) })
	})
}

// ── Custom ValkeyInstance ─────────────────────────────────────────────────────

func testCustomValkeyInstance(t *testing.T, vkNs *terrak8s.KubectlOptions) {
	t.Helper()

	// Create
	waitSyncedAndReady(t, vkNs, ValkeyInstanceKind, ValkeyCustomName, 120, 10*time.Second)
	if t.Failed() {
		return
	}

	rgName, err := getFirstByLabel(t, vkNs, ValkeyReplicationGroupKind, ValkeyCustomName)
	require.NoError(t, err)
	require.NotEmpty(t, rgName)

	// Read: verify custom spec fields propagated to the provider resource
	require.Equal(t, "cache.t4g.medium", getField(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.nodeType"))
	require.Equal(t, "8.2", getField(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.engineVersion"))
	require.Equal(t, "2", getField(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.numCacheClusters"))
	require.Equal(t, "3", getField(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.snapshotRetentionLimit"))
}

// ── Cleanup ───────────────────────────────────────────────────────────────────

func cleanupValkey(t *testing.T, cluster, argocd *terrak8s.KubectlOptions) {
	if t.Failed() {
		return
	}
	vkNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, ValkeyNamespaceName)

	cleanupDeleteParallel(t, vkNs, ValkeyInstanceKind, ValkeyCustomName)

	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocd, "delete", "application", ValkeyApplicationName, "--ignore-not-found")
}
