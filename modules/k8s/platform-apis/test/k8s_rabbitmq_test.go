package test

import (
	"context"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

func testRabbitMQ(t *testing.T, ctx context.Context, cluster, argocd *terrak8s.KubectlOptions) {
	mqNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, RabbitMQNamespaceName)
	defer cleanupRabbitMQ(t, cluster, argocd)

	if ctx.Err() != nil {
		return
	}
	applyFile(t, cluster, "./templates/rabbitmq_test_application.yaml")
	syncWithRetry(t, argocd, RabbitMQApplicationName)
	if ctx.Err() != nil {
		return
	}

	t.Run("broker", func(t *testing.T) {
		t.Run("RabbitMQLifecycle", func(t *testing.T) { testRabbitMQLifecycle(t, mqNs) })
	})

	if t.Failed() {
		return
	}

	t.Run("configuration", func(t *testing.T) {
		t.Run("RabbitMQConfigLifecycle", func(t *testing.T) { testRabbitMQConfigLifecycle(t, mqNs) })
	})
}

// testRabbitMQLifecycle drives a single RabbitMQBroker through provisioning and asserts the AWS
// resources it composes.
func testRabbitMQLifecycle(t *testing.T, mqNs *terrak8s.KubectlOptions) {
	t.Helper()

	// SecurityGroup is fast; the credentials secret is created before the broker is provisioned.
	waitSyncedAndReadyByLabel(t, mqNs, SecurityGroupKind, RabbitMQBrokerName, 60, 10*time.Second)
	waitResourceExists(t, mqNs, "secret", RabbitMQCredentialsSecretName, 60, 10*time.Second)
	if t.Failed() {
		return
	}

	// Credentials secret carries the admin username/password used to bootstrap the broker.
	require.NotEmpty(t, getField(t, mqNs, "secret", RabbitMQCredentialsSecretName, ".data.username"),
		"credentials secret username should be populated")
	require.NotEmpty(t, getField(t, mqNs, "secret", RabbitMQCredentialsSecretName, ".data.password"),
		"credentials secret password should be populated")

	// Broker provisioning is slow - allow a long window for Synced+Ready.
	brokerName := waitSyncedAndReadyByLabel(t, mqNs, RabbitMQAwsBrokerKind, RabbitMQBrokerName, 120, 15*time.Second)
	require.NotEmpty(t, brokerName)
	if t.Failed() {
		return
	}

	// Broker spec must reflect what was specified on the composite.
	require.Equal(t, "mq.m7g.medium", getField(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".spec.forProvider.hostInstanceType"))
	require.Equal(t, RabbitMQStartVersion, getField(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".spec.forProvider.engineVersion"))
	require.Equal(t, "RabbitMQ", getField(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".spec.forProvider.engineType"))
	require.Equal(t, "SINGLE_INSTANCE", getField(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".spec.forProvider.deploymentMode"))

	// The connection secret is written by writeConnectionSecretToRef once the broker is ready.
	waitResourceExists(t, mqNs, "secret", RabbitMQConnectionSecretName, 60, 10*time.Second)
	require.NotEmpty(t, getField(t, mqNs, "secret", RabbitMQConnectionSecretName, ".data"),
		"connection secret should be populated")

	// Composite status must be populated once the broker is ready.
	require.NotEmpty(t, getField(t, mqNs, RabbitMQBrokerKind, RabbitMQBrokerName, ".status.amazonMQBrokerID"),
		"composite amazonMQBrokerID should be populated")
}

// testRabbitMQConfigLifecycle exercises the broker Configuration lifecycle: created & wired at the
// starting engine version, edited in place on a data change (new revision), then on an engine-version
// bump a new version-scoped Configuration is created, wired to the broker, and the old one removed.
func testRabbitMQConfigLifecycle(t *testing.T, mqNs *terrak8s.KubectlOptions) {
	t.Helper()

	brokerName := waitSyncedAndReadyByLabel(t, mqNs, RabbitMQAwsBrokerKind, RabbitMQBrokerName, 120, 15*time.Second)
	require.NotEmpty(t, brokerName)
	if t.Failed() {
		return
	}

	// 1-2: Configuration for the starting version is created, Ready, and wired to the broker.
	oldConfig := waitSyncedAndReadyByLabelWhere(t, mqNs, RabbitMQAwsConfigKind, RabbitMQBrokerName, ".spec.forProvider.engineVersion", RabbitMQStartVersion, 120, 15*time.Second)
	require.NotEmpty(t, oldConfig)
	waitFieldEquals(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".spec.forProvider.configuration.idRef.name", oldConfig, 60, 10*time.Second)
	waitFieldNonEmpty(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".spec.forProvider.configuration.revision", 120, 15*time.Second)
	if t.Failed() {
		return
	}

	// 3-4: Changing data edits the same Configuration in place (new revision), still wired to the broker.
	// Assert the data actually propagated (revision numbers are not deterministic - the provider may
	// create extra revisions on create/normalization) and the Configuration name did not change.
	patchResource(t, mqNs, RabbitMQBrokerKind, RabbitMQBrokerName, `{"spec":{"configuration":{"data":"consumer_timeout = 3600000\n"}}}`)
	waitFieldContains(t, mqNs, RabbitMQAwsConfigKind, oldConfig, ".status.atProvider.data", "3600000", 120, 15*time.Second)
	require.Equal(t, oldConfig, getField(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".spec.forProvider.configuration.idRef.name"),
		"Configuration must be edited in place, not recreated, on a data change")
	waitFieldNonEmpty(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".spec.forProvider.configuration.revision", 120, 15*time.Second)
	if t.Failed() {
		return
	}

	// 5-6: Bumping the engine version creates a new version-scoped Configuration (engineVersion immutable).
	patchResource(t, mqNs, RabbitMQBrokerKind, RabbitMQBrokerName, `{"spec":{"engineVersion":"`+RabbitMQUpgradeVersion+`"}}`)
	newConfig := waitSyncedAndReadyByLabelWhere(t, mqNs, RabbitMQAwsConfigKind, RabbitMQBrokerName, ".spec.forProvider.engineVersion", RabbitMQUpgradeVersion, 120, 15*time.Second)
	require.NotEmpty(t, newConfig)
	require.NotEqual(t, oldConfig, newConfig, "a new Configuration must be created for the new engine version")

	// 7: The broker upgrades to the new version and switches its association to the new Configuration.
	//    The engine upgrade is slow, so allow a long window.
	waitFieldEquals(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".spec.forProvider.configuration.idRef.name", newConfig, 240, 15*time.Second)
	waitFieldEquals(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".spec.forProvider.engineVersion", RabbitMQUpgradeVersion, 240, 15*time.Second)
	if t.Failed() {
		return
	}

	// 8: Once the broker has switched, the old-version Configuration is removed.
	cleanupWaitGone(t, mqNs, RabbitMQAwsConfigKind, oldConfig, 120)

	// Post-upgrade the broker stays healthy: connection secret and composite status still hold.
	require.NotEmpty(t, getField(t, mqNs, "secret", RabbitMQConnectionSecretName, ".data"),
		"connection secret should still be populated after the upgrade")
	require.NotEmpty(t, getField(t, mqNs, RabbitMQBrokerKind, RabbitMQBrokerName, ".status.amazonMQBrokerID"),
		"composite amazonMQBrokerID should still be populated after the upgrade")
	waitFieldEquals(t, mqNs, RabbitMQBrokerKind, RabbitMQBrokerName, ".status.engineVersion", RabbitMQUpgradeVersion, 240, 15*time.Second)
}

func cleanupRabbitMQ(t *testing.T, cluster, argocd *terrak8s.KubectlOptions) {
	if t.Failed() {
		return // leave resources in place for debugging
	}
	mqNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, RabbitMQNamespaceName)
	cleanupDeleteParallel(t, mqNs, RabbitMQBrokerKind, 180, RabbitMQBrokerName)

	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocd, "delete", "application", RabbitMQApplicationName, "--ignore-not-found")
}
