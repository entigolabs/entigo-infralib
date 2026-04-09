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

func kafkaK8sSecretName() string { return KafkaClusterName + "-" + KafkaUserName + "-k8s" }
func kafkaAWSSecretName() string { return KafkaClusterName + "-" + KafkaUserName }
func kafkaAWSSecVerName() string { return KafkaClusterName + "-" + KafkaUserName + "-version" }
func kafkaAWSSecPolName() string { return KafkaClusterName + "-" + KafkaUserName + "-policy" }
func kafkaSCRAMName() string     { return KafkaClusterName + "-" + KafkaUserName + "-scram" }

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

	applyFile(t, cluster, "./templates/kafka_test_application.yaml")
	setKafkaApplicationClusterARN(t, argocd, clusterARN)
	syncWithRetry(t, argocd, KafkaApplicationName)
	if ctx.Err() != nil {
		return
	}

	t.Run("MSKObserver", func(t *testing.T) { testMSKObserver(t, cluster) })
	if t.Failed() {
		return
	}

	t.Run("Topic", func(t *testing.T) { testKafkaTopic(t, kfNs) })
	if t.Failed() {
		return
	}

	t.Run("User", func(t *testing.T) { testKafkaUser(t, kfNs) })
}

func setKafkaClusterApplicationClusterARN(t *testing.T, argocd *terrak8s.KubectlOptions, arn string) {
	t.Helper()
	values := fmt.Sprintf("targetRevision: '*.*.*-0'\nclusterName: %q\nclusterARN: %q\nrenderCompositions: false\n", KafkaClusterName, arn)
	valuesJSON, err := json.Marshal(values)
	require.NoError(t, err, "marshal helm values")
	patch := fmt.Sprintf(`{"spec":{"source":{"helm":{"values":%s}}}}`, string(valuesJSON))
	patchResource(t, argocd, "application", KafkaClusterApplicationName, patch)
}

func setKafkaApplicationClusterARN(t *testing.T, argocd *terrak8s.KubectlOptions, arn string) {
	t.Helper()
	values := fmt.Sprintf("targetRevision: '*.*.*-0'\nclusterName: %q\nclusterARN: %q\nrenderCompositions: true\n", KafkaClusterName, arn)
	valuesJSON, err := json.Marshal(values)
	require.NoError(t, err, "marshal helm values")
	patch := fmt.Sprintf(`{"spec":{"source":{"helm":{"values":%s}}}}`, string(valuesJSON))
	patchResource(t, argocd, "application", KafkaApplicationName, patch)
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
	require.NotEmpty(t, getField(t, cluster, KafkaMSKKind, KafkaMSKObserverName, ".status.brokersscram"), "MSK brokersscram should be populated")
	require.Equal(t, KafkaMSKObserverName,
		getField(t, cluster, KafkaMSKKind, KafkaMSKObserverName, ".status.providerConfig"))

	provCfgName, err := getFirstByLabel(t, cluster, KafkaClusterProvCfgKind, KafkaMSKObserverName)
	require.NoError(t, err)
	require.NotEmpty(t, provCfgName, "ClusterProviderConfig should exist for MSK observer")
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

func testKafkaUser(t *testing.T, kfNs *terrak8s.KubectlOptions) {
	t.Helper()

	// Create
	waitSyncedAndReady(t, kfNs, KafkaUserKind, KafkaUserName, 60, 10*time.Second)
	if t.Failed() {
		return
	}

	t.Run("Secrets", func(t *testing.T) {
		t.Run("K8sSecret", func(t *testing.T) {
			t.Parallel()
			waitResourceExists(t, kfNs, "secret", kafkaK8sSecretName(), 30, 10*time.Second)
			require.NotEmpty(t, getField(t, kfNs, "secret", kafkaK8sSecretName(), ".data.secretString"))
		})
		t.Run("AWSSecret", func(t *testing.T) {
			t.Parallel()
			waitSyncedAndReady(t, kfNs, KafkaAWSSecKind, kafkaAWSSecretName(), 60, 10*time.Second)
			require.Equal(t, "AmazonMSK_"+KafkaClusterName+"-"+KafkaUserName,
				getField(t, kfNs, KafkaAWSSecKind, kafkaAWSSecretName(), ".spec.forProvider.name"),
				"AWS Secret name should use AmazonMSK_ prefix")
		})
		t.Run("SecretVersion", func(t *testing.T) {
			t.Parallel()
			waitSyncedAndReady(t, kfNs, KafkaAWSSecVerKind, kafkaAWSSecVerName(), 60, 10*time.Second)
			require.Equal(t, KafkaUserName,
				getField(t, kfNs, KafkaAWSSecVerKind, kafkaAWSSecVerName(), ".spec.forProvider.secretIdSelector.matchLabels.kafka-user"))
			require.Equal(t, KafkaClusterName,
				getField(t, kfNs, KafkaAWSSecVerKind, kafkaAWSSecVerName(), ".spec.forProvider.secretIdSelector.matchLabels.msk-cluster"))
		})
		t.Run("SecretPolicy", func(t *testing.T) {
			t.Parallel()
			waitSyncedAndReady(t, kfNs, KafkaAWSSecPolKind, kafkaAWSSecPolName(), 60, 10*time.Second)
		})
		t.Run("SCRAMAssociation", func(t *testing.T) {
			t.Parallel()
			waitSyncedAndReady(t, kfNs, KafkaSCRAMKind, kafkaSCRAMName(), 60, 10*time.Second)
			require.Equal(t, kafkaAWSSecretName(),
				getField(t, kfNs, KafkaSCRAMKind, kafkaSCRAMName(), ".spec.forProvider.secretArnRef.name"))
			require.NotEmpty(t, getField(t, kfNs, KafkaSCRAMKind, kafkaSCRAMName(), ".spec.forProvider.clusterArn"))
		})
	})
	if t.Failed() {
		return
	}

	// Consumer group ACLs
	t.Run("ConsumerGroupACLs", func(t *testing.T) {
		for _, cg := range []string{"alpha", "gamma"} {
			cg := cg
			t.Run(cg, func(t *testing.T) {
				t.Parallel()
				aclName := KafkaUserName + "-" + cg + "-cg"
				waitSyncedAndReady(t, kfNs, KafkaACLKind, aclName, 30, 10*time.Second)
				require.Equal(t, "Group", getField(t, kfNs, KafkaACLKind, aclName, ".spec.forProvider.resourceType"))
				require.Equal(t, cg, getField(t, kfNs, KafkaACLKind, aclName, ".spec.forProvider.resourceName"))
				require.Equal(t, "User:"+KafkaUserName, getField(t, kfNs, KafkaACLKind, aclName, ".spec.forProvider.resourcePrincipal"))
				require.Equal(t, "Read", getField(t, kfNs, KafkaACLKind, aclName, ".spec.forProvider.resourceOperation"))
				require.Equal(t, "Allow", getField(t, kfNs, KafkaACLKind, aclName, ".spec.forProvider.resourcePermissionType"))
			})
		}
	})

	// Topic ACLs
	t.Run("TopicACLs", func(t *testing.T) {
		t.Run("Read", func(t *testing.T) {
			t.Parallel()
			aclName := KafkaTopicName + "-" + KafkaUserName + "-read"
			waitSyncedAndReady(t, kfNs, KafkaACLKind, aclName, 30, 10*time.Second)
			require.Equal(t, "Topic", getField(t, kfNs, KafkaACLKind, aclName, ".spec.forProvider.resourceType"))
			require.Equal(t, KafkaTopicName, getField(t, kfNs, KafkaACLKind, aclName, ".spec.forProvider.resourceName"))
			require.Equal(t, "User:"+KafkaUserName, getField(t, kfNs, KafkaACLKind, aclName, ".spec.forProvider.resourcePrincipal"))
			require.Equal(t, "Read", getField(t, kfNs, KafkaACLKind, aclName, ".spec.forProvider.resourceOperation"))
			require.Equal(t, "Allow", getField(t, kfNs, KafkaACLKind, aclName, ".spec.forProvider.resourcePermissionType"))
		})
		t.Run("Write", func(t *testing.T) {
			t.Parallel()
			aclName := KafkaTopicName + "-" + KafkaUserName + "-write"
			waitSyncedAndReady(t, kfNs, KafkaACLKind, aclName, 30, 10*time.Second)
			require.Equal(t, "Topic", getField(t, kfNs, KafkaACLKind, aclName, ".spec.forProvider.resourceType"))
			require.Equal(t, KafkaTopicName, getField(t, kfNs, KafkaACLKind, aclName, ".spec.forProvider.resourceName"))
			require.Equal(t, "Write", getField(t, kfNs, KafkaACLKind, aclName, ".spec.forProvider.resourceOperation"))
		})
	})

	// Update: add a new consumer group → new ACL must appear
	patchResource(t, kfNs, KafkaUserKind, KafkaUserName, `{"spec":{"consumerGroups":["alpha","gamma","delta"]}}`)
	newACL := KafkaUserName + "-delta-cg"
	waitSyncedAndReady(t, kfNs, KafkaACLKind, newACL, 30, 10*time.Second)
	require.Equal(t, "delta", getField(t, kfNs, KafkaACLKind, newACL, ".spec.forProvider.resourceName"))
}

func cleanupKafka(t *testing.T, cluster, argocd *terrak8s.KubectlOptions) {
	if t.Failed() {
		return
	}
	kfNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, KafkaNamespaceName)
	kfClusterNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, KafkaClusterNamespaceName)

	cleanupDeleteParallel(t, kfNs, KafkaUserKind, KafkaUserName)
	cleanupDeleteParallel(t, kfNs, KafkaTopicKind, KafkaTopicName)

	cleanupDeleteAndWait(t, cluster, KafkaMSKKind, KafkaMSKObserverName, 30)

	_, _ = terrak8s.RunKubectlAndGetOutputE(t, kfClusterNs, "delete", KafkaMSKClusterKind, KafkaClusterName, "--ignore-not-found")
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, kfClusterNs, "delete", "securitygroup.ec2.aws.upbound.io", KafkaSecurityGroupName, "--ignore-not-found")

	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocd, "delete", "application", KafkaApplicationName, "--ignore-not-found")
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocd, "delete", "application", KafkaClusterApplicationName, "--ignore-not-found")
}
