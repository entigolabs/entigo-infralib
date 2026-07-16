package test

import (
	"context"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

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
		t.Run("ValkeyLifecycle", func(t *testing.T) { testValkeyLifecycle(t, vkNs) })
	})
}

// testValkeyLifecycle drives a single ValkeyInstance through every spec/engineVersion/
// parameterGroupParameters combination worth covering, since each transition is a real (slow) AWS
// ElastiCache change and provisioning a separate cluster per case would multiply e2e cost for no
// extra coverage.
//
// The instance starts at a deliberately low engineVersion ("7.2", see the helm template). AWS
// ElastiCache only supports engine_version upgrades, never downgrades - the provider explicitly
// refuses a downgrade as a ForceNew ("cannot change the value of the argument engine_version...").
// Starting low keeps every later engineVersion transition in this test a valid upgrade, regardless
// of what AWS's own "no version specified" default happens to be at the time.
func testValkeyLifecycle(t *testing.T, vkNs *terrak8s.KubectlOptions) {
	t.Helper()

	waitSyncedAndReady(t, vkNs, ValkeyInstanceKind, ValkeyLifecycleName, 120, 10*time.Second)
	if t.Failed() {
		return
	}

	rgName, err := getFirstByLabel(t, vkNs, ValkeyReplicationGroupKind, ValkeyLifecycleName)
	require.NoError(t, err)
	require.NotEmpty(t, rgName)

	require.Equal(t, "cache.t4g.medium", getField(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.nodeType"))
	require.Equal(t, "2", getField(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.numCacheClusters"))
	require.Equal(t, "3", getField(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.snapshotRetentionLimit"))
	require.Equal(t, "7.2", getField(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.engineVersion"))

	_, err = getFirstByLabel(t, vkNs, ValkeyParameterGroupKind, ValkeyLifecycleName)
	require.Error(t, err, "no ParameterGroup should exist while parameterGroupParameters is unset")

	// Case: engineVersion removed -> must not force any change to the already-applied version.
	// (Same code path as engineVersion never being set at all: the field just isn't included in
	// the desired ReplicationGroup spec, so the provider leaves the current value untouched.)
	patchResource(t, vkNs, ValkeyInstanceKind, ValkeyLifecycleName, `{"spec":{"engineVersion":null}}`)
	waitFieldEquals(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.engineVersion", "7.2", 60, 10*time.Second)

	actualFamily := "valkey7"

	patchResource(t, vkNs, ValkeyInstanceKind, ValkeyLifecycleName, `{"spec":{"parameterGroupParameters":{"notify-keyspace-events":"Ex"}}}`)

	pgName := waitSyncedAndReadyByLabel(t, vkNs, ValkeyParameterGroupKind, ValkeyLifecycleName, 60, 10*time.Second)
	require.NotEmpty(t, pgName)
	require.Equal(t, actualFamily, getField(t, vkNs, ValkeyParameterGroupKind, pgName, ".spec.forProvider.family"))
	require.Equal(t, "notify-keyspace-events", getField(t, vkNs, ValkeyParameterGroupKind, pgName, ".spec.forProvider.parameter[0].name"))
	require.Equal(t, "Ex", getField(t, vkNs, ValkeyParameterGroupKind, pgName, ".spec.forProvider.parameter[0].value"))
	waitFieldEquals(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.parameterGroupName", pgName, 60, 10*time.Second)

	// Changing a parameter value (same family) must update the existing ParameterGroup in place -
	// parameter/value is Optional on the provider, not ForceNew, so no delete/recreate should happen.
	patchResource(t, vkNs, ValkeyInstanceKind, ValkeyLifecycleName, `{"spec":{"parameterGroupParameters":{"notify-keyspace-events":"KEA"}}}`)

	waitFieldEquals(t, vkNs, ValkeyParameterGroupKind, pgName, ".spec.forProvider.parameter[0].value", "KEA", 60, 10*time.Second)
	require.Equal(t, pgName, getField(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.parameterGroupName"),
		"ParameterGroup should update in place, not be recreated under a different name")

	// engineVersion changed to a different major version (7.2 -> 8.2, a valid upgrade) -> family is
	// immutable on the AWS ParameterGroup and AWS refuses to delete one still referenced by a
	// ReplicationGroup, so the old and new ParameterGroup must coexist until the ReplicationGroup's
	// observed status (not just its desired spec) confirms the switch - only then can the old one
	// actually be deleted.
	newVersion, newFamily := "8.2", "valkey8"
	patchResource(t, vkNs, ValkeyInstanceKind, ValkeyLifecycleName, `{"spec":{"engineVersion":"`+newVersion+`"}}`)

	recreatedPgName := waitSyncedAndReadyByLabelWhere(t, vkNs, ValkeyParameterGroupKind, ValkeyLifecycleName, ".spec.forProvider.family", newFamily, 60, 10*time.Second)
	require.NotEqual(t, pgName, recreatedPgName)
	// The observed parameterGroupName only flips once AWS finishes the 7.2 -> 8.2 major engine upgrade,
	// which is a slow ElastiCache operation that routinely exceeds 20 min. Allow up to 45 min.
	waitFieldEquals(t, vkNs, ValkeyReplicationGroupKind, rgName, ".status.atProvider.parameterGroupName", recreatedPgName, 270, 10*time.Second)
	waitFieldEquals(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.parameterGroupName", recreatedPgName, 60, 10*time.Second)
	waitFieldEquals(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.engineVersion", newVersion, 60, 10*time.Second)

	cleanupWaitGone(t, vkNs, ValkeyParameterGroupKind, pgName, 30)

	// parameterGroupParameters removed -> ParameterGroup deleted, ReplicationGroup reverts
	// explicitly to the engine default (parameterGroupName is Optional+Computed on the provider, so
	// merely omitting it would leave the previous custom group name untouched).
	patchResource(t, vkNs, ValkeyInstanceKind, ValkeyLifecycleName, `{"spec":{"parameterGroupParameters":null}}`)

	cleanupWaitGone(t, vkNs, ValkeyParameterGroupKind, recreatedPgName, 30)
	waitFieldEquals(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.parameterGroupName", "default."+newFamily, 60, 10*time.Second)
}

func cleanupValkey(t *testing.T, cluster, argocd *terrak8s.KubectlOptions) {
	if t.Failed() {
		return
	}
	vkNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, ValkeyNamespaceName)

	cleanupDeleteParallel(t, vkNs, ValkeyInstanceKind, 30, ValkeyLifecycleName)

	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocd, "delete", "application", ValkeyApplicationName, "--ignore-not-found")
}
