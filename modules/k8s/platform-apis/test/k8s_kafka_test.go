package test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

func testKafka(t *testing.T, ctx context.Context, cluster, argocd *terrak8s.KubectlOptions) {
	kfNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, KafkaNamespaceName)
	kfClusterNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, KafkaClusterNamespaceName)
	defer cleanupKafka(t, cluster, argocd)

	if ctx.Err() != nil {
		return
	}

	applyFile(t, cluster, "./templates/kafka_cluster_application.yaml")
	syncWithRetry(t, argocd, KafkaClusterApplicationName)
	if ctx.Err() != nil {
		return
	}

	waitSyncedAndReady(t, kfClusterNs, KafkaMSKClusterKind, KafkaClusterName, 120, 30*time.Second)
	if t.Failed() {
		return
	}

	clusterARN := getField(t, kfClusterNs, KafkaMSKClusterKind, KafkaClusterName, ".status.atProvider.arn")
	require.NotEmpty(t, clusterARN, "MSK cluster ARN must be populated in status")
	if t.Failed() {
		return
	}

	setKafkaClusterApplicationClusterARN(t, argocd, clusterARN)
	syncWithRetry(t, argocd, KafkaClusterApplicationName)
	if ctx.Err() != nil {
		return
	}

	t.Run("MSKObserver", func(t *testing.T) { testMSKObserver(t, cluster) })
	if t.Failed() {
		return
	}

	t.Run("Topic", func(t *testing.T) { testKafkaTopic(t, kfNs) })
}

func setKafkaClusterApplicationClusterARN(t *testing.T, argocd *terrak8s.KubectlOptions, arn string) {
	t.Helper()
	values := fmt.Sprintf("targetRevision: '*.*.*-0'\nclusterName: %q\nclusterARN: %q\nrenderCompositions: false\n", KafkaClusterName, arn)
	valuesJSON, err := json.Marshal(values)
	require.NoError(t, err, "marshal helm values")
	patch := fmt.Sprintf(`{"spec":{"source":{"helm":{"values":%s}}}}`, string(valuesJSON))
	patchResource(t, argocd, "application", KafkaClusterApplicationName, patch)
}

func testMSKObserver(t *testing.T, cluster *terrak8s.KubectlOptions) {
	t.Helper()

	waitSyncedAndReady(t, cluster, KafkaMSKKind, KafkaMSKObserverName, 30, 10*time.Second)
	if t.Failed() {
		return
	}

	require.NotEmpty(t, getField(t, cluster, KafkaMSKKind, KafkaMSKObserverName, ".status.arn"), "MSK arn should be populated")
	require.NotEmpty(t, getField(t, cluster, KafkaMSKKind, KafkaMSKObserverName, ".status.region"), "MSK region should be populated")
	require.NotEmpty(t, getField(t, cluster, KafkaMSKKind, KafkaMSKObserverName, ".status.brokers"), "MSK brokers should be populated")
	provCfgName, err := getFirstByLabel(t, cluster, KafkaClusterProvCfgKind, KafkaMSKObserverName)
	require.NoError(t, err)
	require.NotEmpty(t, provCfgName, "ClusterProviderConfig should exist for MSK observer")
	require.Equal(t, provCfgName,
		getField(t, cluster, KafkaMSKKind, KafkaMSKObserverName, ".status.providerConfig"),
		"MSK providerConfig status should match ClusterProviderConfig name")
}

func testKafkaTopic(t *testing.T, kfNs *terrak8s.KubectlOptions) {
	t.Helper()

	// Create
	waitSyncedAndReady(t, kfNs, KafkaTopicKind, KafkaTopicName, 30, 10*time.Second)
	if t.Failed() {
		return
	}

	provTopicName, err := getFirstByLabel(t, kfNs, KafkaProvTopicKind, KafkaTopicName)
	require.NoError(t, err)
	require.NotEmpty(t, provTopicName)
	waitSyncedAndReady(t, kfNs, KafkaProvTopicKind, provTopicName, 30, 10*time.Second)

	// Read
	require.Equal(t, KafkaTopicPartitions, getField(t, kfNs, KafkaProvTopicKind, provTopicName, ".spec.forProvider.partitions"))
	require.Equal(t, KafkaTopicReplicationFactor, getField(t, kfNs, KafkaProvTopicKind, provTopicName, ".spec.forProvider.replicationFactor"))

	// Update: increase partitions (Kafka only allows increases)
	patchResource(t, kfNs, KafkaTopicKind, KafkaTopicName, `{"spec":{"partitions":`+KafkaTopicUpdatedPartitions+`}}`)
	waitFieldEquals(t, kfNs, KafkaProvTopicKind, provTopicName, ".spec.forProvider.partitions", KafkaTopicUpdatedPartitions, 30, 10*time.Second)
}

func cleanupKafka(t *testing.T, cluster, argocd *terrak8s.KubectlOptions) {
	if t.Failed() {
		return
	}
	kfNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, KafkaNamespaceName)
	kfClusterNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, KafkaClusterNamespaceName)

	cleanupDeleteParallel(t, kfNs, KafkaTopicKind, KafkaTopicName)

	cleanupDeleteAndWait(t, cluster, KafkaMSKKind, KafkaMSKObserverName, 30)

	_, _ = terrak8s.RunKubectlAndGetOutputE(t, kfClusterNs, "delete", KafkaMSKClusterKind, KafkaClusterName, "--ignore-not-found")
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, kfClusterNs, "delete", "securitygrouprule.ec2.aws.upbound.io", KafkaSecurityGroupIngressName, "--ignore-not-found")
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, kfClusterNs, "delete", "securitygroup.ec2.aws.upbound.io", KafkaSecurityGroupName, "--ignore-not-found")

	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocd, "delete", "application", KafkaClusterApplicationName, "--ignore-not-found")
}
