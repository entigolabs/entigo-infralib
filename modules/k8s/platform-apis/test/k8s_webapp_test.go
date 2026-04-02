package test

import (
	"context"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

// ── Orchestrator ──────────────────────────────────────────────────────────────

func testWebApp(t *testing.T, ctx context.Context, cluster, argocd *terrak8s.KubectlOptions) {
	waNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, WebAppNamespaceName)
	defer cleanupWebApp(t, cluster, argocd)

	if ctx.Err() != nil {
		return
	}
	applyFile(t, cluster, "./templates/webapp_test_application.yaml")
	syncWithRetry(t, argocd, WebAppApplicationName)
	if ctx.Err() != nil {
		return
	}

	t.Run("WebApp", func(t *testing.T) { testWebAppResource(t, waNs) })
}

// ── WebApp ────────────────────────────────────────────────────────────────────

func testWebAppResource(t *testing.T, waNs *terrak8s.KubectlOptions) {
	t.Helper()

	// Create: wait for composite to be Synced+Ready (function sets this once Deployment is Available)
	waitSyncedAndReady(t, waNs, WebAppKind, WebAppName, 60, 10*time.Second)
	if t.Failed() {
		return
	}

	// Sub-resources
	t.Run("SubResources", func(t *testing.T) {
		t.Run("Deployment", func(t *testing.T) {
			t.Parallel()
			waitResourceExists(t, waNs, "deployment", WebAppDeploymentName, 30, 10*time.Second)
		})
		t.Run("Service", func(t *testing.T) {
			t.Parallel()
			waitResourceExists(t, waNs, "service", WebAppServiceName, 30, 10*time.Second)
		})
		t.Run("Secret", func(t *testing.T) {
			t.Parallel()
			waitResourceExists(t, waNs, "secret", WebAppSecretName, 30, 10*time.Second)
		})
	})
	if t.Failed() {
		return
	}

	// Read: verify spec fields propagated to the Deployment
	require.Equal(t, "docker.io/nginx:alpine",
		getField(t, waNs, "deployment", WebAppDeploymentName, ".spec.template.spec.containers[0].image"))
	require.Equal(t, "1",
		getField(t, waNs, "deployment", WebAppDeploymentName, ".spec.replicas"))
	require.Equal(t, "80",
		getField(t, waNs, "service", WebAppServiceName, ".spec.ports[0].port"))

	// Update: scale replicas and verify the Deployment follows
	patchResource(t, waNs, WebAppKind, WebAppName, `{"spec":{"replicas":2}}`)
	waitFieldEquals(t, waNs, "deployment", WebAppDeploymentName, ".spec.replicas", "2", 30, 10*time.Second)
}

// ── Cleanup ───────────────────────────────────────────────────────────────────

func cleanupWebApp(t *testing.T, cluster, argocd *terrak8s.KubectlOptions) {
	if t.Failed() {
		return
	}
	waNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, WebAppNamespaceName)

	cleanupDeleteAndWait(t, waNs, WebAppKind, WebAppName, 30)

	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocd, "delete", "application", WebAppApplicationName, "--ignore-not-found")
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, cluster, "delete", "namespace", WebAppNamespaceName, "--ignore-not-found", "--wait=true")
}
