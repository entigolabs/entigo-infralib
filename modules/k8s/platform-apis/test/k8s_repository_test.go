package test

import (
	"context"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

func testRepository(t *testing.T, ctx context.Context, cluster, argocd *terrak8s.KubectlOptions) {
	repoNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, RepositoryNamespaceName)
	defer cleanupRepository(t, cluster, argocd)

	if ctx.Err() != nil {
		return
	}
	applyFile(t, cluster, "./templates/repository_test_application.yaml")
	syncWithRetry(t, argocd, RepositoryApplicationName)
	if ctx.Err() != nil {
		return
	}

	t.Run("repositories", func(t *testing.T) {
		t.Run("MinimalRepository", func(t *testing.T) { t.Parallel(); testMinimalRepository(t, repoNs) })
		t.Run("NamedRepository", func(t *testing.T) { t.Parallel(); testNamedRepository(t, repoNs) })
	})
}

func testMinimalRepository(t *testing.T, repoNs *terrak8s.KubectlOptions) {
	t.Helper()

	// Create
	waitSyncedAndReady(t, repoNs, RepositoryKind, RepositoryMinimalName, 60, 10*time.Second)
	if t.Failed() {
		return
	}

	ecrName, err := getFirstByLabel(t, repoNs, ECRRepositoryKind, RepositoryMinimalName)
	require.NoError(t, err)
	require.NotEmpty(t, ecrName)
	waitSyncedAndReady(t, repoNs, ECRRepositoryKind, ecrName, 60, 10*time.Second)

	require.Equal(t, RepositoryMinimalName,
		getField(t, repoNs, ECRRepositoryKind, ecrName, `.metadata.annotations.crossplane\.io/external-name`))
	require.NotEmpty(t, getField(t, repoNs, RepositoryKind, RepositoryMinimalName, ".status.repositoryUri"),
		"repositoryUri should be populated once ECR repo is ready")
	require.Equal(t, "test",
		getField(t, repoNs, ECRRepositoryKind, ecrName, `.metadata.labels.tags\.entigo\.com/tag`))
	require.Equal(t, "test",
		getField(t, repoNs, ECRRepositoryKind, ecrName, `.spec.forProvider.tags.tag`))
	require.Equal(t, "eztest",
		getField(t, repoNs, ECRRepositoryKind, ecrName, `.spec.forProvider.tags.etag`))
	require.Equal(t, "eutest",
		getField(t, repoNs, ECRRepositoryKind, ecrName, `.spec.forProvider.tags.eutag`))
	require.Equal(t, "zutest",
		getField(t, repoNs, ECRRepositoryKind, ecrName, `.spec.forProvider.tags.zutag`))
	require.Equal(t, "antest",
		getField(t, repoNs, ECRRepositoryKind, ecrName, `.spec.forProvider.tags.antag`))
}

func testNamedRepository(t *testing.T, repoNs *terrak8s.KubectlOptions) {
	t.Helper()

	// Create
	waitSyncedAndReady(t, repoNs, RepositoryKind, RepositoryNamedName, 60, 10*time.Second)
	if t.Failed() {
		return
	}

	ecrName, err := getFirstByLabel(t, repoNs, ECRRepositoryKind, RepositoryNamedName)
	require.NoError(t, err)
	require.NotEmpty(t, ecrName)
	waitSyncedAndReady(t, repoNs, ECRRepositoryKind, ecrName, 60, 10*time.Second)

	require.Equal(t, RepositoryNamedExternalName,
		getField(t, repoNs, ECRRepositoryKind, ecrName, `.metadata.annotations.crossplane\.io/external-name`))
	require.Equal(t, RepositoryNamedECRName,
		getField(t, repoNs, RepositoryKind, RepositoryNamedName, ".spec.name"))
	require.Equal(t, RepositoryNamedPath,
		getField(t, repoNs, RepositoryKind, RepositoryNamedName, ".spec.path"))
	require.NotEmpty(t, getField(t, repoNs, RepositoryKind, RepositoryNamedName, ".status.repositoryUri"))
	require.Equal(t, "ztest",
		getField(t, repoNs, ECRRepositoryKind, ecrName, `.metadata.labels.tags\.entigo\.com/tag`))
	require.Equal(t, "ztest",
		getField(t, repoNs, ECRRepositoryKind, ecrName, `.spec.forProvider.tags.tag`))
	require.Equal(t, "eztest",
		getField(t, repoNs, ECRRepositoryKind, ecrName, `.spec.forProvider.tags.etag`))
	require.Equal(t, "eutest",
		getField(t, repoNs, ECRRepositoryKind, ecrName, `.spec.forProvider.tags.eutag`))
	require.Equal(t, "zutest",
		getField(t, repoNs, ECRRepositoryKind, ecrName, `.spec.forProvider.tags.zutag`))

	// Update: name and path are immutable — patch must be rejected
	_, err = terrak8s.RunKubectlAndGetOutputE(t, repoNs, "patch", RepositoryKind, RepositoryNamedName,
		"--type", "merge", "-p", `{"spec":{"name":"changed-name"}}`)
	require.Error(t, err, "patching immutable spec.name should be rejected")
}

func cleanupRepository(t *testing.T, cluster, argocd *terrak8s.KubectlOptions) {
	if t.Failed() {
		return
	}
	repoNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, RepositoryNamespaceName)

	cleanupDeleteParallel(t, repoNs, RepositoryKind, RepositoryMinimalName, RepositoryNamedName)

	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocd, "delete", "application", RepositoryApplicationName, "--ignore-not-found")
}
