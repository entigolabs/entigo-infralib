package test

import (
	"context"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

// testRabbitMQ drives two brokers side by side. The private broker carries the slow path: provisioning,
// then the engine-version bump with its Configuration switch (one broker reboot). The public broker
// carries the public-access assertions and the in-place Configuration data change (its own reboot), so
// that reboot no longer sits in front of the private broker's upgrade. Each broker is deleted as soon
// as its own chain is done; the Application goes once both are finished.
func testRabbitMQ(t *testing.T, ctx context.Context, cluster, argocd *terrak8s.KubectlOptions) {
	mqNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, RabbitMQNamespaceName)
	t.Cleanup(func() { cleanupRabbitMQApplication(t, argocd) })

	if ctx.Err() != nil {
		return
	}
	applyFile(t, cluster, "./templates/rabbitmq_test_application.yaml")
	syncWithRetry(t, argocd, RabbitMQApplicationName)
	if ctx.Err() != nil {
		return
	}

	t.Run("Private", func(t *testing.T) {
		t.Parallel()
		defer cleanupRabbitMQBroker(t, mqNs, RabbitMQBrokerName)
		t.Run("RabbitMQLifecycle", func(t *testing.T) { testRabbitMQLifecycle(t, mqNs) })
		if t.Failed() {
			return
		}
		t.Run("RabbitMQConfigUpgrade", func(t *testing.T) { testRabbitMQConfigUpgrade(t, mqNs) })
	})

	t.Run("Public", func(t *testing.T) {
		t.Parallel()
		defer cleanupRabbitMQBroker(t, mqNs, RabbitMQPublicBrokerName)
		t.Run("RabbitMQPublicAccessImmutable", func(t *testing.T) { testRabbitMQPublicAccessImmutable(t, mqNs) })
		t.Run("RabbitMQPublicBroker", func(t *testing.T) { testRabbitMQPublicBroker(t, mqNs) })
		if t.Failed() {
			return
		}
		t.Run("RabbitMQConfigDataChange", func(t *testing.T) {
			testRabbitMQConfigDataChange(t, mqNs, RabbitMQPublicBrokerName, RabbitMQUpgradeVersion)
		})
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

	for _, port := range []string{"5671", "443", "15671"} {
		require.NotEmpty(t, waitSyncedAndReadyByLabelWhere(t, mqNs, SecurityGroupRuleKind, RabbitMQBrokerName,
			".spec.forProvider.fromPort", port, 60, 10*time.Second), "no ingress rule for port %s", port)
	}
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

// waitRabbitMQConfigWired waits for the Configuration of the given engine version to be Ready and wired
// to the broker composed for the composite, and returns both names.
func waitRabbitMQConfigWired(t *testing.T, mqNs *terrak8s.KubectlOptions, composite, engineVersion string) (brokerName, configName string) {
	t.Helper()

	brokerName = waitSyncedAndReadyByLabel(t, mqNs, RabbitMQAwsBrokerKind, composite, 120, 15*time.Second)
	require.NotEmpty(t, brokerName)
	if t.Failed() {
		return "", ""
	}

	configName = waitSyncedAndReadyByLabelWhere(t, mqNs, RabbitMQAwsConfigKind, composite, ".spec.forProvider.engineVersion", engineVersion, 120, 15*time.Second)
	require.NotEmpty(t, configName)
	waitFieldEquals(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".spec.forProvider.configuration.idRef.name", configName, 60, 10*time.Second)
	waitFieldNonEmpty(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".spec.forProvider.configuration.revision", 120, 15*time.Second)
	return brokerName, configName
}

// testRabbitMQConfigDataChange exercises the in-place half of the Configuration lifecycle on the broker
// composed for the given composite: the Configuration for its engine version is created and wired, and a
// data change edits that same Configuration in place (new revision) instead of recreating it. The
// provider applies the new revision with a broker reboot; the test does not wait for it.
func testRabbitMQConfigDataChange(t *testing.T, mqNs *terrak8s.KubectlOptions, composite, engineVersion string) {
	t.Helper()

	brokerName, configName := waitRabbitMQConfigWired(t, mqNs, composite, engineVersion)
	if t.Failed() {
		return
	}

	// Assert the data actually propagated (revision numbers are not deterministic - the provider may
	// create extra revisions on create/normalization) and the Configuration name did not change.
	patchResource(t, mqNs, RabbitMQBrokerKind, composite, `{"spec":{"configuration":{"data":"consumer_timeout = 3600000\n"}}}`)
	waitFieldContains(t, mqNs, RabbitMQAwsConfigKind, configName, ".status.atProvider.data", "3600000", 120, 15*time.Second)
	require.Equal(t, configName, getField(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".spec.forProvider.configuration.idRef.name"),
		"Configuration must be edited in place, not recreated, on a data change")
	waitFieldNonEmpty(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".spec.forProvider.configuration.revision", 120, 15*time.Second)
}

// testRabbitMQConfigUpgrade exercises the engine-version half of the Configuration lifecycle on the
// private broker: bumping the engine version creates a new version-scoped Configuration, the broker
// upgrades and switches to it in one update, and the old Configuration is removed afterwards.
func testRabbitMQConfigUpgrade(t *testing.T, mqNs *terrak8s.KubectlOptions) {
	t.Helper()

	brokerName, oldConfig := waitRabbitMQConfigWired(t, mqNs, RabbitMQBrokerName, RabbitMQStartVersion)
	if t.Failed() {
		return
	}

	// Bumping the engine version creates a new version-scoped Configuration (engineVersion immutable).
	patchResource(t, mqNs, RabbitMQBrokerKind, RabbitMQBrokerName, `{"spec":{"engineVersion":"`+RabbitMQUpgradeVersion+`"}}`)
	newConfig := waitSyncedAndReadyByLabelWhere(t, mqNs, RabbitMQAwsConfigKind, RabbitMQBrokerName, ".spec.forProvider.engineVersion", RabbitMQUpgradeVersion, 120, 15*time.Second)
	require.NotEmpty(t, newConfig)
	require.NotEqual(t, oldConfig, newConfig, "a new Configuration must be created for the new engine version")

	// The broker upgrades to the new version and switches its association to the new Configuration.
	// The engine upgrade is slow, so allow a long window.
	waitFieldEquals(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".spec.forProvider.configuration.idRef.name", newConfig, 240, 15*time.Second)
	waitFieldEquals(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".spec.forProvider.engineVersion", RabbitMQUpgradeVersion, 240, 15*time.Second)
	waitFieldEquals(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".status.atProvider.engineVersion", RabbitMQUpgradeVersion, 240, 15*time.Second)
	waitSyncedAndReady(t, mqNs, RabbitMQAwsBrokerKind, brokerName, 60, 15*time.Second)
	if t.Failed() {
		return
	}

	// Once the broker has switched, the old-version Configuration is removed.
	cleanupWaitGone(t, mqNs, RabbitMQAwsConfigKind, oldConfig, 120)

	// Post-upgrade the broker stays healthy: connection secret and composite status still hold.
	require.NotEmpty(t, getField(t, mqNs, "secret", RabbitMQConnectionSecretName, ".data"),
		"connection secret should still be populated after the upgrade")
	require.NotEmpty(t, getField(t, mqNs, RabbitMQBrokerKind, RabbitMQBrokerName, ".status.amazonMQBrokerID"),
		"composite amazonMQBrokerID should still be populated after the upgrade")
	waitFieldEquals(t, mqNs, RabbitMQBrokerKind, RabbitMQBrokerName, ".status.engineVersion", RabbitMQUpgradeVersion, 240, 15*time.Second)
}

func testRabbitMQPublicAccessImmutable(t *testing.T, mqNs *terrak8s.KubectlOptions) {
	t.Helper()

	before := getField(t, mqNs, RabbitMQBrokerKind, RabbitMQBrokerName, ".spec.publiclyAccessible")
	require.Equal(t, "false", before, "broker should start out private")

	out, err := terrak8s.RunKubectlAndGetOutputE(t, mqNs, "patch", RabbitMQBrokerKind, RabbitMQBrokerName,
		"--type", "merge", "-p", `{"spec":{"publiclyAccessible":true}}`)
	require.Error(t, err, "switching publiclyAccessible on an existing broker must be rejected")
	require.Contains(t, out, "publiclyAccessible is immutable")

	require.Equal(t, before, getField(t, mqNs, RabbitMQBrokerKind, RabbitMQBrokerName, ".spec.publiclyAccessible"),
		"a rejected patch must leave the broker private")
}

// testRabbitMQPublicBroker drives a second, publiclyAccessible broker through provisioning. Amazon MQ
// rejects CreateBroker for a RabbitMQ broker that sets both publiclyAccessible and securityGroups, so no
// SecurityGroup stack is composed. Subnets stay attached - the subnetIds assertions are what catch it if
// Amazon MQ turns out to reject those on a public broker too.
func testRabbitMQPublicBroker(t *testing.T, mqNs *terrak8s.KubectlOptions) {
	t.Helper()

	waitResourceExists(t, mqNs, "secret", RabbitMQPublicCredentialsSecretName, 60, 10*time.Second)

	brokerName := waitSyncedAndReadyByLabel(t, mqNs, RabbitMQAwsBrokerKind, RabbitMQPublicBrokerName, 120, 15*time.Second)
	require.NotEmpty(t, brokerName)
	if t.Failed() {
		return
	}

	require.Empty(t, getField(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".spec.forProvider.securityGroupRefs"),
		"a public broker must not reference any security group")
	require.Empty(t, getField(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".status.atProvider.securityGroups"),
		"a public broker must not report any security group")
	require.NotEmpty(t, getField(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".spec.forProvider.subnetIds"),
		"subnets must stay attached on a public broker")
	require.NotEmpty(t, getField(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".status.atProvider.subnetIds"),
		"Amazon MQ must accept subnetIds on a public broker")

	// No SecurityGroup or SecurityGroupRule is composed for this broker at all.
	waitNoResourcesByLabel(t, mqNs, SecurityGroupKind, RabbitMQPublicBrokerName, 3, 5*time.Second)
	waitNoResourcesByLabel(t, mqNs, SecurityGroupRuleKind, RabbitMQPublicBrokerName, 3, 5*time.Second)

	waitResourceExists(t, mqNs, "secret", RabbitMQPublicConnectionSecretName, 60, 10*time.Second)
	require.NotEmpty(t, getField(t, mqNs, "secret", RabbitMQPublicConnectionSecretName, ".data"),
		"connection secret should be populated for a public broker")
	waitFieldEquals(t, mqNs, RabbitMQBrokerKind, RabbitMQPublicBrokerName, ".status.publiclyAccessible", "true", 60, 10*time.Second)
	require.NotEmpty(t, getField(t, mqNs, RabbitMQBrokerKind, RabbitMQPublicBrokerName, ".status.amazonMQBrokerID"),
		"composite amazonMQBrokerID should be populated")
}

// cleanupRabbitMQBroker deletes one broker composite as soon as its own chain is done.
func cleanupRabbitMQBroker(t *testing.T, mqNs *terrak8s.KubectlOptions, name string) {
	if t.Failed() {
		return // leave resources in place for debugging
	}
	cleanupDeleteParallel(t, mqNs, RabbitMQBrokerKind, 180, name)
}

// cleanupRabbitMQApplication removes the ArgoCD Application once both broker chains have finished.
func cleanupRabbitMQApplication(t *testing.T, argocd *terrak8s.KubectlOptions) {
	if t.Failed() {
		return
	}
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocd, "delete", "application", RabbitMQApplicationName, "--ignore-not-found")
}
