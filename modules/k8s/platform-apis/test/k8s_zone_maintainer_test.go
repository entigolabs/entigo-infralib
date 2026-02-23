package test

import (
	"fmt"
	"os"
	"testing"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

const MaintainerTestZoneName = "maintainer-test-zone"

func testZoneMaintainerPolicies(t *testing.T, clusterOptions *terrak8s.KubectlOptions, argocdOptions *terrak8s.KubectlOptions) {
	t.Helper()
	maintainerOptions := setupRoleOptions(t, clusterOptions, "MAINTAINER_AWS_ACCESS_KEY_ID", "MAINTAINER_AWS_SECRET_ACCESS_KEY")
	maintainerArgocdOptions := terrak8s.NewKubectlOptions(maintainerOptions.ContextName, maintainerOptions.ConfigPath, argocdOptions.Namespace)

	// Sequential: own zone must exist before namespace-in-own-zone sub-test
	t.Run("can-create-and-delete-zone", func(t *testing.T) {
		testMaintainerZoneLifecycle(t, clusterOptions, maintainerOptions)
	})

	t.Run("can-create-namespace-in-zone-a", func(t *testing.T) {
		t.Parallel()
		testMaintainerCanCreateNamespaceInZone(t, clusterOptions, maintainerOptions, "maintainer-test-ns-a", ZoneAName)
	})
	t.Run("can-create-namespace-in-zone-b", func(t *testing.T) {
		t.Parallel()
		testMaintainerCanCreateNamespaceInZone(t, clusterOptions, maintainerOptions, "maintainer-test-ns-b", ZoneBName)
	})
	t.Run("cannot-delete-namespace-in-zone-a", func(t *testing.T) {
		t.Parallel()
		_, err := terrak8s.RunKubectlAndGetOutputE(t, maintainerOptions, "delete", "namespace", ZoneAAppsNamespace)
		require.Error(t, err, "maintainer should not be able to delete namespace in zone A")
	})
	t.Run("cannot-update-namespace-in-zone-a", func(t *testing.T) {
		t.Parallel()
		_, err := terrak8s.RunKubectlAndGetOutputE(t, maintainerOptions, "label", "namespace", ZoneAAppsNamespace, "test-label=value", "--overwrite")
		require.Error(t, err, "maintainer should not be able to update namespace in zone A")
	})
	// test-zone namespace (created by testNamespaceCreated) is labeled zone A but not in zone A spec.namespaces,
	// so Kyverno platform-apis-zone-deletion-check-namespaces blocks the deletion
	t.Run("cannot-delete-zone-a", func(t *testing.T) {
		t.Parallel()
		_, err := terrak8s.RunKubectlAndGetOutputE(t, maintainerOptions, "delete", "zone.tenancy.entigo.com", ZoneAName)
		require.Error(t, err, "maintainer should not be able to delete zone A (has unmanaged namespaces)")
	})
	t.Run("cannot-create-app-in-zone-a-project", func(t *testing.T) {
		t.Parallel()
		_, err := terrak8s.RunKubectlAndGetOutputE(t, maintainerArgocdOptions, "apply", "-f", "./templates/zone_test_a_apps.yaml")
		require.Error(t, err, "maintainer should not be able to create/update ArgoCD application in zone A project")
	})
}

func testMaintainerZoneLifecycle(t *testing.T, clusterOptions *terrak8s.KubectlOptions, maintainerOptions *terrak8s.KubectlOptions) {
	t.Helper()
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "delete", "zone.tenancy.entigo.com", MaintainerTestZoneName, "--ignore-not-found")
	defer terrak8s.RunKubectl(t, clusterOptions, "delete", "zone.tenancy.entigo.com", MaintainerTestZoneName, "--ignore-not-found")

	_, err := terrak8s.RunKubectlAndGetOutputE(t, maintainerOptions, "apply", "-f", "./templates/zone_test_maintainer_zone.yaml")
	require.NoError(t, err, "maintainer should be able to create a zone")

	name, err := terrak8s.RunKubectlAndGetOutputE(t, maintainerOptions, "get", "zone.tenancy.entigo.com", MaintainerTestZoneName, "-o", "jsonpath={.metadata.name}")
	require.NoError(t, err)
	require.Equal(t, MaintainerTestZoneName, name)

	t.Run("can-create-namespace-in-own-zone", func(t *testing.T) {
		testMaintainerCanCreateNamespaceInZone(t, clusterOptions, maintainerOptions, "maintainer-test-ns-own", MaintainerTestZoneName)
	})

	// namespace is already cleaned up by defer inside testMaintainerCanCreateNamespaceInZone,
	// so the deletion check (which blocks deletion when unmanaged namespaces exist) will not fire
	_, err = terrak8s.RunKubectlAndGetOutputE(t, maintainerOptions, "delete", "zone.tenancy.entigo.com", MaintainerTestZoneName)
	require.NoError(t, err, "maintainer should be able to delete own zone")
}

func testMaintainerCanCreateNamespaceInZone(t *testing.T, clusterOptions *terrak8s.KubectlOptions, maintainerOptions *terrak8s.KubectlOptions, nsName, zoneName string) {
	t.Helper()
	nsYAML := fmt.Sprintf("apiVersion: v1\nkind: Namespace\nmetadata:\n  name: %s\n  labels:\n    tenancy.entigo.com/zone: %s\n", nsName, zoneName)

	tmpFile, err := os.CreateTemp("", "ns-*.yaml")
	require.NoError(t, err)
	tmpPath := tmpFile.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	_, err = tmpFile.WriteString(nsYAML)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	// maintainer cannot delete namespaces, admin cleans up
	defer terrak8s.RunKubectl(t, clusterOptions, "delete", "namespace", nsName, "--ignore-not-found")

	_, err = terrak8s.RunKubectlAndGetOutputE(t, maintainerOptions, "apply", "-f", tmpPath)
	require.NoError(t, err, "maintainer should be able to create namespace %s in zone %s", nsName, zoneName)

	result, err := terrak8s.RunKubectlAndGetOutputE(t, maintainerOptions, "get", "namespace", nsName, "-o", "jsonpath={.metadata.name}")
	require.NoError(t, err)
	require.Equal(t, nsName, result)
}
