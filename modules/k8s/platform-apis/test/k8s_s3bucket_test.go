package test

import (
	"context"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

// ── Orchestrator ──────────────────────────────────────────────────────────────

func testS3Bucket(t *testing.T, ctx context.Context, cluster, argocd *terrak8s.KubectlOptions) {
	s3Ns := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, S3BucketNamespaceName)
	defer cleanupS3Bucket(t, cluster, argocd)

	if ctx.Err() != nil {
		return
	}
	applyFile(t, cluster, "./templates/s3bucket_test_application.yaml")
	syncWithRetry(t, argocd, S3BucketApplicationName)
	if ctx.Err() != nil {
		return
	}

	t.Run("buckets", func(t *testing.T) {
		t.Run("MinimalS3Bucket", func(t *testing.T) { t.Parallel(); testMinimalS3Bucket(t, s3Ns) })
		t.Run("VersionedS3Bucket", func(t *testing.T) { t.Parallel(); testVersionedS3Bucket(t, s3Ns) })
	})
}

// ── Minimal S3 Bucket ─────────────────────────────────────────────────────────

func testMinimalS3Bucket(t *testing.T, s3Ns *terrak8s.KubectlOptions) {
	t.Helper()

	// Create
	waitSyncedAndReady(t, s3Ns, S3BucketKind, S3MinimalName, 90, 10*time.Second)
	if t.Failed() {
		return
	}

	// Sub-resources must be Synced+Ready
	t.Run("SubResources", func(t *testing.T) {
		t.Run("Bucket", func(t *testing.T) {
			t.Parallel()
			waitSyncedAndReadyByLabel(t, s3Ns, S3BucketAwsKind, S3MinimalName, 90, 10*time.Second)
		})
		t.Run("IAMUser", func(t *testing.T) {
			t.Parallel()
			waitSyncedAndReadyByLabel(t, s3Ns, S3IAMUserKind, S3MinimalName, 60, 10*time.Second)
		})
		t.Run("IAMRole", func(t *testing.T) {
			t.Parallel()
			waitSyncedAndReadyByLabel(t, s3Ns, S3IAMRoleKind, S3MinimalName, 60, 10*time.Second)
		})
		t.Run("IAMPolicy", func(t *testing.T) {
			t.Parallel()
			waitSyncedAndReadyByLabel(t, s3Ns, S3IAMPolicyKind, S3MinimalName, 60, 10*time.Second)
		})
		t.Run("SecretsManagerSecret", func(t *testing.T) {
			t.Parallel()
			waitSyncedAndReadyByLabel(t, s3Ns, S3SecretsManagerSecretKind, S3MinimalName, 60, 10*time.Second)
		})
	})
	if t.Failed() {
		return
	}

	// Read: status fields must be populated
	require.NotEmpty(t, getField(t, s3Ns, S3BucketKind, S3MinimalName, ".status.s3Uri"),
		"status.s3Uri should be populated")
	require.NotEmpty(t, getField(t, s3Ns, S3BucketKind, S3MinimalName, ".status.region"),
		"status.region should be populated")
	require.Equal(t, "true",
		getField(t, s3Ns, S3BucketKind, S3MinimalName, ".status.blockPublicAclsEnabled"))
	require.Equal(t, "true",
		getField(t, s3Ns, S3BucketKind, S3MinimalName, ".status.blockPublicPolicyEnabled"))

	// ServiceAccount must exist (createServiceAccount defaults to true)
	waitResourceExists(t, s3Ns, "serviceaccount", S3MinimalName, 30, 10*time.Second)

	// Update: enable versioning and verify it propagates to status
	patchResource(t, s3Ns, S3BucketKind, S3MinimalName, `{"spec":{"enableVersioning":true}}`)
	waitFieldEquals(t, s3Ns, S3BucketKind, S3MinimalName, ".status.versioningEnabled", "true", 60, 10*time.Second)
}

// ── Versioned S3 Bucket ───────────────────────────────────────────────────────

func testVersionedS3Bucket(t *testing.T, s3Ns *terrak8s.KubectlOptions) {
	t.Helper()

	// Create
	waitSyncedAndReady(t, s3Ns, S3BucketKind, S3VersionedName, 90, 10*time.Second)
	if t.Failed() {
		return
	}

	// Read: versioning must be enabled from the start
	require.Equal(t, "true",
		getField(t, s3Ns, S3BucketKind, S3VersionedName, ".status.versioningEnabled"))
}

// ── Cleanup ───────────────────────────────────────────────────────────────────

func cleanupS3Bucket(t *testing.T, cluster, argocd *terrak8s.KubectlOptions) {
	if t.Failed() {
		return
	}
	s3Ns := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, S3BucketNamespaceName)

	cleanupDeleteParallel(t, s3Ns, S3BucketKind, S3MinimalName, S3VersionedName)

	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocd, "delete", "application", S3BucketApplicationName, "--ignore-not-found")
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, cluster, "delete", "namespace", S3BucketNamespaceName, "--ignore-not-found", "--wait=true")
}
