package test

import (
	"fmt"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
)

const (
	KafkaTopicMRKind = "topic.topic.kafka.m.crossplane.io"
)

func runKafkaTopicTests(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	waitSyncedAndReady(t, namespaceOptions, KafkaTopicKind, KafkaTopicName, 90, 10*time.Second)

	t.Run("sub-resources", func(t *testing.T) {
		t.Run("kafka-topic-mr", func(t *testing.T) {
			testKafkaTopicMRSyncedAndReady(t, namespaceOptions)
		})
	})

	testTopicFieldsVerified(t, namespaceOptions)
}

func testKafkaTopicMRSyncedAndReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for Kafka Topic MR (composite=%s)", KafkaTopicName), 60, 10*time.Second, func() (string, error) {
		name, err := getFirstByLabel(t, namespaceOptions, KafkaTopicMRKind, KafkaTopicName)
		if err != nil {
			return "", err
		}
		if name == "" {
			return "", fmt.Errorf("no Kafka Topic MR found for composite=%s", KafkaTopicName)
		}
		for _, condType := range []string{"Synced", "Ready"} {
			status, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", KafkaTopicMRKind, name, "-o",
				fmt.Sprintf(`jsonpath={.status.conditions[?(@.type=="%s")].status}`, condType))
			if err != nil {
				return "", err
			}
			if status != "True" {
				return "", fmt.Errorf("Kafka Topic MR '%s': %s=%s", name, condType, status)
			}
		}
		return name, nil
	})
	require.NoError(t, err, "Kafka Topic MR for '%s' failed to become Synced and Ready", KafkaTopicName)
}

func testTopicFieldsVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	t.Helper()
	topicMRName, err := getFirstByLabel(t, namespaceOptions, KafkaTopicMRKind, KafkaTopicName)
	require.NoError(t, err, "failed to find Kafka Topic MR")
	require.NotEmpty(t, topicMRName, "no Kafka Topic MR found for composite '%s'", KafkaTopicName)

	partitions, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", KafkaTopicMRKind, topicMRName, "-o", "jsonpath={.spec.forProvider.partitions}")
	require.NoError(t, err, "failed to get partitions")
	require.Equal(t, "3", partitions, "Kafka Topic MR '%s' partitions mismatch", topicMRName)

	replicationFactor, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", KafkaTopicMRKind, topicMRName, "-o", "jsonpath={.spec.forProvider.replicationFactor}")
	require.NoError(t, err, "failed to get replicationFactor")
	require.Equal(t, "2", replicationFactor, "Kafka Topic MR '%s' replicationFactor mismatch", topicMRName)
}
