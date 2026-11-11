package test

import (
	"context"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

// testValkey drives two ValkeyInstances side by side. The Lifecycle instance carries the slow path:
// provisioning, then the engine-version bump with the ParameterGroup family switch that comes with it.
// The Parameters instance carries the parameterGroupParameters set / change in place / remove
// lifecycle, so those ElastiCache modifications no longer sit in front of the upgrade. Each instance is
// deleted as soon as its own chain is done; the Application goes once both are finished.
func testValkey(t *testing.T, ctx context.Context, cluster, argocd *terrak8s.KubectlOptions) {
	vkNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, ValkeyNamespaceName)
	t.Cleanup(func() { cleanupValkeyApplication(t, argocd) })

	if ctx.Err() != nil {
		return
	}
	applyFile(t, cluster, "./templates/valkey_test_application.yaml")
	syncWithRetry(t, argocd, ValkeyApplicationName)
	if ctx.Err() != nil {
		return
	}

	t.Run("Lifecycle", func(t *testing.T) {
		t.Parallel()
		defer cleanupValkeyInstance(t, vkNs, ValkeyLifecycleName)
		t.Run("ValkeyLifecycle", func(t *testing.T) { testValkeyLifecycle(t, vkNs) })
	})

	t.Run("Parameters", func(t *testing.T) {
		t.Parallel()
		defer cleanupValkeyInstance(t, vkNs, ValkeyParametersName)
		t.Run("ValkeyParameters", func(t *testing.T) { testValkeyParameters(t, vkNs) })
	})
}

// waitValkeyReplicationGroup waits for the ValkeyInstance to be Ready and returns its ReplicationGroup,
// asserting the spec the helm template set on it.
func waitValkeyReplicationGroup(t *testing.T, vkNs *terrak8s.KubectlOptions, instance string) string {
	t.Helper()

	waitSyncedAndReady(t, vkNs, ValkeyInstanceKind, instance, 120, 10*time.Second)
	if t.Failed() {
		return ""
	}

	rgName, err := getFirstByLabel(t, vkNs, ValkeyReplicationGroupKind, instance)
	require.NoError(t, err)
	require.NotEmpty(t, rgName)

	require.Equal(t, "cache.t4g.medium", getField(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.nodeType"))
	require.Equal(t, "2", getField(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.numCacheClusters"))
	require.Equal(t, "3", getField(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.snapshotRetentionLimit"))
	require.Equal(t, "7.2", getField(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.engineVersion"))
	return rgName
}

// testValkeyLifecycle drives the Lifecycle instance through the engineVersion transitions, since each
// one is a real (slow) ElastiCache change.
//
// The instance starts at a deliberately low engineVersion ("7.2", see the helm template). AWS
// ElastiCache only supports engine_version upgrades, never downgrades - the provider explicitly
// refuses a downgrade as a ForceNew ("cannot change the value of the argument engine_version...").
// Starting low keeps every later engineVersion transition in this test a valid upgrade, regardless
// of what AWS's own "no version specified" default happens to be at the time.
func testValkeyLifecycle(t *testing.T, vkNs *terrak8s.KubectlOptions) {
	t.Helper()

	rgName := waitValkeyReplicationGroup(t, vkNs, ValkeyLifecycleName)
	if t.Failed() {
		return
	}

	// The template sets parameterGroupParameters, so a ParameterGroup for the starting family is
	// composed and wired from the start - the engine bump below has one to recreate.
	actualFamily := "valkey7"
	pgName := waitSyncedAndReadyByLabel(t, vkNs, ValkeyParameterGroupKind, ValkeyLifecycleName, 60, 10*time.Second)
	require.NotEmpty(t, pgName)
	require.Equal(t, actualFamily, getField(t, vkNs, ValkeyParameterGroupKind, pgName, ".spec.forProvider.family"))
	require.Equal(t, "notify-keyspace-events", getField(t, vkNs, ValkeyParameterGroupKind, pgName, ".spec.forProvider.parameter[0].name"))
	require.Equal(t, "Ex", getField(t, vkNs, ValkeyParameterGroupKind, pgName, ".spec.forProvider.parameter[0].value"))
	waitFieldEquals(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.parameterGroupName", pgName, 60, 10*time.Second)

	// Case: engineVersion removed -> must not force any change to the already-applied version.
	// (Same code path as engineVersion never being set at all: the field just isn't included in
	// the desired ReplicationGroup spec, so the provider leaves the current value untouched.)
	patchResource(t, vkNs, ValkeyInstanceKind, ValkeyLifecycleName, `{"spec":{"engineVersion":null}}`)
	waitFieldEquals(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.engineVersion", "7.2", 60, 10*time.Second)

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
}

// testValkeyParameters drives the Parameters instance through the parameterGroupParameters lifecycle:
// no ParameterGroup while unset, one created and wired when set, updated in place on a value change,
// and deleted again with the ReplicationGroup reverting to the engine default when removed.
func testValkeyParameters(t *testing.T, vkNs *terrak8s.KubectlOptions) {
	t.Helper()

	rgName := waitValkeyReplicationGroup(t, vkNs, ValkeyParametersName)
	if t.Failed() {
		return
	}

	_, err := getFirstByLabel(t, vkNs, ValkeyParameterGroupKind, ValkeyParametersName)
	require.Error(t, err, "no ParameterGroup should exist while parameterGroupParameters is unset")

	actualFamily := "valkey7"

	patchResource(t, vkNs, ValkeyInstanceKind, ValkeyParametersName, `{"spec":{"parameterGroupParameters":{"notify-keyspace-events":"Ex"}}}`)

	pgName := waitSyncedAndReadyByLabel(t, vkNs, ValkeyParameterGroupKind, ValkeyParametersName, 60, 10*time.Second)
	require.NotEmpty(t, pgName)
	require.Equal(t, actualFamily, getField(t, vkNs, ValkeyParameterGroupKind, pgName, ".spec.forProvider.family"))
	require.Equal(t, "notify-keyspace-events", getField(t, vkNs, ValkeyParameterGroupKind, pgName, ".spec.forProvider.parameter[0].name"))
	require.Equal(t, "Ex", getField(t, vkNs, ValkeyParameterGroupKind, pgName, ".spec.forProvider.parameter[0].value"))
	waitFieldEquals(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.parameterGroupName", pgName, 60, 10*time.Second)

	// Changing a parameter value (same family) must update the existing ParameterGroup in place -
	// parameter/value is Optional on the provider, not ForceNew, so no delete/recreate should happen.
	patchResource(t, vkNs, ValkeyInstanceKind, ValkeyParametersName, `{"spec":{"parameterGroupParameters":{"notify-keyspace-events":"KEA"}}}`)
	waitFieldEquals(t, vkNs, ValkeyParameterGroupKind, pgName, ".spec.forProvider.parameter[0].value", "KEA", 60, 10*time.Second)
	require.Equal(t, pgName, getField(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.parameterGroupName"),
		"ParameterGroup should update in place, not be recreated under a different name")

	// parameterGroupParameters removed -> ParameterGroup deleted, ReplicationGroup reverts
	// explicitly to the engine default (parameterGroupName is Optional+Computed on the provider, so
	// merely omitting it would leave the previous custom group name untouched).
	patchResource(t, vkNs, ValkeyInstanceKind, ValkeyParametersName, `{"spec":{"parameterGroupParameters":null}}`)

	cleanupWaitGone(t, vkNs, ValkeyParameterGroupKind, pgName, 30)
	waitFieldEquals(t, vkNs, ValkeyReplicationGroupKind, rgName, ".spec.forProvider.parameterGroupName", "default."+actualFamily, 60, 10*time.Second)
}

// cleanupValkeyInstance deletes one ValkeyInstance as soon as its own chain is done.
func cleanupValkeyInstance(t *testing.T, vkNs *terrak8s.KubectlOptions, name string) {
	if t.Failed() {
		return
	}
	cleanupDeleteParallel(t, vkNs, ValkeyInstanceKind, 30, name)
}

// cleanupValkeyApplication removes the ArgoCD Application once both instance chains have finished.
func cleanupValkeyApplication(t *testing.T, argocd *terrak8s.KubectlOptions) {
	if t.Failed() {
		return
	}
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocd, "delete", "application", ValkeyApplicationName, "--ignore-not-found")
}
