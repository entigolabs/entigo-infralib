package test

import (
	"fmt"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

type secretStoreData struct {
	Name, Namespace, Store, Source, RemoteKey, Zone, Value string
}

func testZoneSecretStore(t *testing.T, cluster *terrak8s.KubectlOptions) {
	t.Run("secret-store", func(t *testing.T) {
		a1 := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, "a1")
		b1 := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, "b1")
		defer cleanupZoneSecretStore(t, a1, b1)

		legacy := getField(t, cluster, "environmentconfig", DatabaseEnvironmentConfig, ".data.esClusterSecretStore")
		require.NotEmpty(t, legacy, "cluster-wide ClusterSecretStore name not found")

		// Ready means ESO got credentials for the Zone role through the Zone's ServiceAccount
		waitReady(t, cluster, ClusterSecretStoreKind, ZoneSecretStoreA, 30, 10*time.Second)
		if t.Failed() {
			return
		}

		// The test environment's cluster-wide store has write access, so it creates the AWS secrets
		applySecretStoreTemplate(t, a1, "secretstore_source_secret.yaml", secretStoreData{
			Name: SecretStoreSourceName, Namespace: "a1", Value: SecretStoreValue,
		})
		pushSecretFixture(t, a1, SecretStorePushA, legacy, SecretStoreRemoteKeyA, ZoneAName)
		pushSecretFixture(t, a1, SecretStorePushB, legacy, SecretStoreRemoteKeyB, ZoneBName)
		pushSecretFixture(t, a1, SecretStorePushUntagged, legacy, SecretStoreRemoteKeyUntagged, "")
		if t.Failed() {
			return
		}

		t.Run("cluster-store-still-syncs", func(t *testing.T) {
			testExternalSecretSyncs(t, a1, SecretStoreLegacyESName, legacy, SecretStoreRemoteKeyUntagged)
		})
		t.Run("zone-store-reads-own-tag", func(t *testing.T) {
			testExternalSecretSyncs(t, a1, SecretStoreZoneESName, ZoneSecretStoreA, SecretStoreRemoteKeyA)
		})
		t.Run("zone-store-denies-untagged", func(t *testing.T) {
			testExternalSecretDenied(t, a1, SecretStoreUntaggedESName, ZoneSecretStoreA, SecretStoreRemoteKeyUntagged)
		})
		t.Run("zone-store-denies-other-zone-tag", func(t *testing.T) {
			testExternalSecretDenied(t, a1, SecretStoreOtherZoneESName, ZoneSecretStoreA, SecretStoreRemoteKeyB)
		})
		t.Run("zone-store-denies-other-zone-namespace", func(t *testing.T) {
			testExternalSecretDenied(t, b1, SecretStoreForeignESName, ZoneSecretStoreA, SecretStoreRemoteKeyA)
		})
		t.Run("switch-store-keeps-secret", func(t *testing.T) {
			testSwitchStore(t, a1, legacy)
		})
	})
}

func applySecretStoreTemplate(t *testing.T, opts *terrak8s.KubectlOptions, template string, data secretStoreData) {
	t.Helper()
	applyFile(t, opts, writeTempYAML(t, renderTemplate(t, "./templates/"+template, data)))
}

// pushSecretFixture creates an AWS secret, tagged entigo:zone=<zone> unless zone is empty.
func pushSecretFixture(t *testing.T, opts *terrak8s.KubectlOptions, name, store, remoteKey, zone string) {
	t.Helper()
	applySecretStoreTemplate(t, opts, "secretstore_pushsecret.yaml", secretStoreData{
		Name: name, Namespace: opts.Namespace, Store: store,
		Source: SecretStoreSourceName, RemoteKey: remoteKey, Zone: zone,
	})
	waitReady(t, opts, PushSecretKind, name, 30, 10*time.Second)
}

func applyExternalSecret(t *testing.T, opts *terrak8s.KubectlOptions, name, store, remoteKey string) {
	t.Helper()
	applySecretStoreTemplate(t, opts, "secretstore_externalsecret.yaml", secretStoreData{
		Name: name, Namespace: opts.Namespace, Store: store, RemoteKey: remoteKey,
	})
}

func testExternalSecretSyncs(t *testing.T, opts *terrak8s.KubectlOptions, name, store, remoteKey string) {
	t.Helper()
	applyExternalSecret(t, opts, name, store, remoteKey)
	waitReady(t, opts, ExternalSecretKind, name, 30, 10*time.Second)
	require.Equal(t, SecretStoreValue, string(terrak8s.GetSecret(t, opts, name).Data["value"]))
}

// testExternalSecretDenied covers both IAM and the store's namespace condition rejecting the read.
func testExternalSecretDenied(t *testing.T, opts *terrak8s.KubectlOptions, name, store, remoteKey string) {
	t.Helper()
	applyExternalSecret(t, opts, name, store, remoteKey)
	waitFieldEquals(t, opts, ExternalSecretKind, name, `.status.conditions[?(@.type=="Ready")].status`, "False", 30, 10*time.Second)
	out, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", "secret", name, "--ignore-not-found", "-o", "jsonpath={.metadata.name}")
	require.NoError(t, err)
	require.Empty(t, out, "Secret %s/%s was written although %s must not read %s there", opts.Namespace, name, store, remoteKey)
}

// testSwitchStore is the step a team takes to adopt the Zone store.
func testSwitchStore(t *testing.T, opts *terrak8s.KubectlOptions, legacy string) {
	t.Helper()
	testExternalSecretSyncs(t, opts, SecretStoreSwitchESName, legacy, SecretStoreRemoteKeyA)
	uid := getField(t, opts, "secret", SecretStoreSwitchESName, ".metadata.uid")
	refreshed := getField(t, opts, ExternalSecretKind, SecretStoreSwitchESName, ".status.refreshTime")

	patchResource(t, opts, ExternalSecretKind, SecretStoreSwitchESName,
		fmt.Sprintf(`{"spec":{"secretStoreRef":{"name":"%s"}}}`, ZoneSecretStoreA))
	waitFieldChanged(t, opts, ExternalSecretKind, SecretStoreSwitchESName, ".status.refreshTime", refreshed, 30, 10*time.Second)
	waitReady(t, opts, ExternalSecretKind, SecretStoreSwitchESName, 30, 10*time.Second)
	require.Equal(t, uid, getField(t, opts, "secret", SecretStoreSwitchESName, ".metadata.uid"), "Secret was recreated when the store changed")
}

// cleanupZoneSecretStore keeps the Namespaces. Deleting the PushSecrets removes the AWS secrets.
func cleanupZoneSecretStore(t *testing.T, a1, b1 *terrak8s.KubectlOptions) {
	cleanupDeleteParallel(t, a1, ExternalSecretKind, 30, SecretStoreLegacyESName, SecretStoreZoneESName,
		SecretStoreUntaggedESName, SecretStoreOtherZoneESName, SecretStoreSwitchESName)
	cleanupDeleteAndWait(t, b1, ExternalSecretKind, SecretStoreForeignESName, 30)
	cleanupDeleteParallel(t, a1, PushSecretKind, 30, SecretStorePushA, SecretStorePushB, SecretStorePushUntagged)
	cleanupDeleteAndWait(t, a1, "secret", SecretStoreSourceName, 15)
}
