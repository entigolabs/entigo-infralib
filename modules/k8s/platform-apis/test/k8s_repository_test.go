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
		t.Run("LifecycleRepository", func(t *testing.T) { t.Parallel(); testLifecycleRepository(t, repoNs) })
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

	// forceDelete is always on so a non-empty repository can never get stuck in deleting.
	require.Equal(t, "true",
		getField(t, repoNs, ECRRepositoryKind, ecrName, ".spec.forProvider.forceDelete"))

	// deletionProtection=true by default; deletion must be rejected by the admission policy.
	require.Equal(t, "true", getField(t, repoNs, RepositoryKind, RepositoryMinimalName, ".spec.deletionProtection"))
	testDeletionRejected(t, repoNs, RepositoryKind, RepositoryMinimalName)

	// No spec.lifecycleRules, so the repository gets the environment config default. AWS rejects a
	// malformed policy document, so reaching Synced is the real assertion here.
	policyName := waitSyncedAndReadyByLabel(t, repoNs, ECRLifecyclePolicyKind, RepositoryMinimalName, 30, 10*time.Second)
	require.Equal(t, RepositoryMinimalName,
		getField(t, repoNs, ECRLifecyclePolicyKind, policyName, ".spec.forProvider.repository"))
	require.JSONEq(t, environmentLifecyclePolicy,
		getField(t, repoNs, ECRLifecyclePolicyKind, policyName, ".spec.forProvider.policy"))
}

// environmentLifecyclePolicy is the default from artifact.environmentConfig.lifecycleRules in
// test/tests/config/aws_biz.yaml.
const environmentLifecyclePolicy = `{"rules":[
	{"rulePriority":1,"description":"Keep 10 latest images tagged with prefix release, DEPLOYED","selection":{"tagStatus":"tagged","tagPrefixList":["release","DEPLOYED"],"countType":"imageCountMoreThan","countNumber":10},"action":{"type":"expire"}},
	{"rulePriority":2,"description":"Keep 30 latest images matching tag pattern *.*.*","selection":{"tagStatus":"tagged","tagPatternList":["*.*.*"],"countType":"imageCountMoreThan","countNumber":30},"action":{"type":"expire"}},
	{"rulePriority":3,"description":"Keep 5 latest images matching tag pattern feature-*, *-cloud","selection":{"tagStatus":"tagged","tagPatternList":["feature-*","*-cloud"],"countType":"imageCountMoreThan","countNumber":5},"action":{"type":"expire"}},
	{"rulePriority":4,"description":"Expire untagged images older than 7 days","selection":{"tagStatus":"untagged","countType":"sinceImagePushed","countUnit":"days","countNumber":7},"action":{"type":"expire"}},
	{"rulePriority":5,"description":"Expire images older than 90 days","selection":{"tagStatus":"any","countType":"sinceImagePushed","countUnit":"days","countNumber":90},"action":{"type":"expire"}}
]}`

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

func testLifecycleRepository(t *testing.T, repoNs *terrak8s.KubectlOptions) {
	t.Helper()

	waitSyncedAndReady(t, repoNs, RepositoryKind, RepositoryLifecycleName, 60, 10*time.Second)
	if t.Failed() {
		return
	}

	ecrName, err := getFirstByLabel(t, repoNs, ECRRepositoryKind, RepositoryLifecycleName)
	require.NoError(t, err)
	require.NotEmpty(t, ecrName)
	waitSyncedAndReady(t, repoNs, ECRRepositoryKind, ecrName, 60, 10*time.Second)

	// The lifecycle policy is sequenced after the repository, so it only appears once that is ready.
	policyName := waitSyncedAndReadyByLabel(t, repoNs, ECRLifecyclePolicyKind, RepositoryLifecycleName, 30, 10*time.Second)
	require.Equal(t, RepositoryLifecycleName,
		getField(t, repoNs, ECRLifecyclePolicyKind, policyName, ".spec.forProvider.repository"))
	// spec.lifecycleRules replaces the environment default entirely.
	require.JSONEq(t, overrideLifecyclePolicy,
		getField(t, repoNs, ECRLifecyclePolicyKind, policyName, ".spec.forProvider.policy"))
}

const overrideLifecyclePolicy = `{"rules":[
	{"rulePriority":1,"description":"Keep 10 latest images tagged with prefix develop","selection":{"tagStatus":"tagged","tagPrefixList":["develop"],"countType":"imageCountMoreThan","countNumber":10},"action":{"type":"expire"}},
	{"rulePriority":2,"description":"Keep 10 latest images matching tag pattern *-cloud","selection":{"tagStatus":"tagged","tagPatternList":["*-cloud"],"countType":"imageCountMoreThan","countNumber":10},"action":{"type":"expire"}},
	{"rulePriority":3,"description":"Expire untagged images older than 7 days","selection":{"tagStatus":"untagged","countType":"sinceImagePushed","countUnit":"days","countNumber":7},"action":{"type":"expire"}},
	{"rulePriority":4,"description":"Expire images older than 90 days","selection":{"tagStatus":"any","countType":"sinceImagePushed","countUnit":"days","countNumber":90},"action":{"type":"expire"}}
]}`

func cleanupRepository(t *testing.T, cluster, argocd *terrak8s.KubectlOptions) {
	if t.Failed() {
		return
	}
	repoNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, RepositoryNamespaceName)

	for _, name := range []string{RepositoryMinimalName, RepositoryNamedName, RepositoryLifecycleName} {
		patchDeletionProtectionIfEnabled(t, repoNs, RepositoryKind, name)
	}

	cleanupDeleteParallel(t, repoNs, RepositoryKind, 30, RepositoryMinimalName, RepositoryNamedName, RepositoryLifecycleName)

	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocd, "delete", "application", RepositoryApplicationName, "--ignore-not-found")
}
