package test

import (
	"testing"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

func testZoneContributorPolicies(t *testing.T, clusterOptions *terrak8s.KubectlOptions, argocdOptions *terrak8s.KubectlOptions) {
	t.Helper()
	contributorOptions := setupRoleOptions(t, clusterOptions, "CONTRIBUTOR_AWS_ACCESS_KEY_ID", "CONTRIBUTOR_AWS_SECRET_ACCESS_KEY")
	contributorArgocdOptions := terrak8s.NewKubectlOptions(contributorOptions.ContextName, contributorOptions.ConfigPath, argocdOptions.Namespace)

	t.Run("can-read-zone-a", func(t *testing.T) {
		t.Parallel()
		name, err := terrak8s.RunKubectlAndGetOutputE(t, contributorOptions, "get", "zone.tenancy.entigo.com", ZoneAName, "-o", "jsonpath={.metadata.name}")
		require.NoError(t, err, "contributor should be able to read zones")
		require.Equal(t, ZoneAName, name)
	})
	t.Run("can-read-namespace-zone-a", func(t *testing.T) {
		t.Parallel()
		name, err := terrak8s.RunKubectlAndGetOutputE(t, contributorOptions, "get", "namespace", ZoneAAppsNamespace, "-o", "jsonpath={.metadata.name}")
		require.NoError(t, err, "contributor should be able to read namespaces")
		require.Equal(t, ZoneAAppsNamespace, name)
	})
	t.Run("cannot-create-zone", func(t *testing.T) {
		t.Parallel()
		defer terrak8s.RunKubectl(t, clusterOptions, "delete", "zone.tenancy.entigo.com", MaintainerTestZoneName, "--ignore-not-found")
		_, err := terrak8s.RunKubectlAndGetOutputE(t, contributorOptions, "apply", "-f", "./templates/zone_test_maintainer_zone.yaml")
		require.Error(t, err, "contributor should not be able to create a zone")
	})
	t.Run("cannot-delete-zone-a", func(t *testing.T) {
		t.Parallel()
		_, err := terrak8s.RunKubectlAndGetOutputE(t, contributorOptions, "delete", "zone.tenancy.entigo.com", ZoneAName)
		require.Error(t, err, "contributor should not be able to delete a zone")
	})
	t.Run("cannot-create-namespace", func(t *testing.T) {
		t.Parallel()
		defer terrak8s.RunKubectl(t, clusterOptions, "delete", "namespace", "contributor-test-ns", "--ignore-not-found")
		_, err := terrak8s.RunKubectlAndGetOutputE(t, contributorOptions, "create", "namespace", "contributor-test-ns")
		require.Error(t, err, "contributor should not be able to create a namespace")
	})
	t.Run("cannot-update-namespace-in-zone-a", func(t *testing.T) {
		t.Parallel()
		_, err := terrak8s.RunKubectlAndGetOutputE(t, contributorOptions, "label", "namespace", ZoneAAppsNamespace, "test-label=value", "--overwrite")
		require.Error(t, err, "contributor should not be able to update a namespace")
	})
	t.Run("cannot-delete-namespace-in-zone-a", func(t *testing.T) {
		t.Parallel()
		_, err := terrak8s.RunKubectlAndGetOutputE(t, contributorOptions, "delete", "namespace", ZoneAAppsNamespace)
		require.Error(t, err, "contributor should not be able to delete a namespace")
	})
	t.Run("cannot-create-app-in-zone-a-project", func(t *testing.T) {
		t.Parallel()
		_, err := terrak8s.RunKubectlAndGetOutputE(t, contributorArgocdOptions, "apply", "-f", "./templates/zone_test_a_apps.yaml")
		require.Error(t, err, "contributor should not be able to create/update ArgoCD application")
	})
}
