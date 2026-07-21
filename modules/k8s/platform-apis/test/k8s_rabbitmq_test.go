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
	require.Equal(t, "4.2", getField(t, mqNs, RabbitMQAwsBrokerKind, brokerName, ".spec.forProvider.engineVersion"))
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

func cleanupRabbitMQ(t *testing.T, cluster, argocd *terrak8s.KubectlOptions) {
	if t.Failed() {
		return // leave resources in place for debugging
	}
	mqNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, RabbitMQNamespaceName)
	cleanupDeleteParallel(t, mqNs, RabbitMQBrokerKind, 180, RabbitMQBrokerName)

	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocd, "delete", "application", RabbitMQApplicationName, "--ignore-not-found")
}
