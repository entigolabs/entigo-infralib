package test

import (
	"context"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

func testCronjob(t *testing.T, ctx context.Context, cluster, argocd *terrak8s.KubectlOptions) {
	cjNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, CronjobNamespaceName)
	defer cleanupCronjob(t, cluster, argocd)

	if ctx.Err() != nil {
		return
	}
	applyFile(t, cluster, "./templates/cronjob_test_application.yaml")
	syncWithRetry(t, argocd, CronjobApplicationName)
	if ctx.Err() != nil {
		return
	}

	t.Run("CronJob", func(t *testing.T) { testCronJobResource(t, cjNs) })
}

func testCronJobResource(t *testing.T, cjNs *terrak8s.KubectlOptions) {
	t.Helper()

	//Create
	waitSyncedAndReady(t, cjNs, CronjobKind, CronjobName, 30, 10*time.Second)
	if t.Failed() {
		return
	}
	waitResourceExists(t, cjNs, BatchCronjobKind, CronjobName, 10, 10*time.Second)
	if t.Failed() {
		return
	}

	// Read: Verify initial spec fields
	require.Equal(t, CronjobInitialSchedule,
		getField(t, cjNs, BatchCronjobKind, CronjobName, ".spec.schedule"))
	require.Equal(t, "Allow",
		getField(t, cjNs, BatchCronjobKind, CronjobName, ".spec.concurrencyPolicy"))
	require.Equal(t, "busybox",
		getField(t, cjNs, BatchCronjobKind, CronjobName, ".spec.jobTemplate.spec.template.spec.containers[0].name"))
	require.Equal(t, "OnFailure",
		getField(t, cjNs, BatchCronjobKind, CronjobName, ".spec.jobTemplate.spec.template.spec.restartPolicy"))

	// Update: change schedule
	patchResource(t, cjNs, CronjobKind, CronjobName, `{"spec":{"schedule":"30 * * * *"}}`)
	waitFieldEquals(t, cjNs, BatchCronjobKind, CronjobName, ".spec.schedule", CronjobUpdatedSchedule, 30, 10*time.Second)

	// Restore original schedule
	patchResource(t, cjNs, CronjobKind, CronjobName, `{"spec":{"schedule":"0 * * * *"}}`)
	waitFieldEquals(t, cjNs, BatchCronjobKind, CronjobName, ".spec.schedule", CronjobInitialSchedule, 30, 10*time.Second)
}

// Delete
func cleanupCronjob(t *testing.T, cluster, argocd *terrak8s.KubectlOptions) {
	if t.Failed() {
		return
	}
	cjNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, CronjobNamespaceName)

	cleanupDeleteAndWait(t, cjNs, CronjobKind, CronjobName, 30)

	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocd, "delete", "application", CronjobApplicationName, "--ignore-not-found")
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, cluster, "delete", "namespace", CronjobNamespaceName, "--ignore-not-found", "--wait=true")
}
