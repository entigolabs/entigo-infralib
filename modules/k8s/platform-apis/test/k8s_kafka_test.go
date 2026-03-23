package test

import (
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
)

const (
	KafkaConfigurationName = "platform-apis-kafka"
	QueueFunctionName      = "platform-apis-queue-fn"
	KafkaNamespaceName     = "test-kafka"
	KafkaApplicationName   = "test-kafka"
	KafkaClusterName       = "test-kafka-cluster"
	MskObserverName        = "test-kafka-cluster-observed"
	KafkaTopicName         = "test-topic-a"
	KafkaUserName          = "test-kafka-user"

	MskAwsClusterKind = "cluster.kafka.aws.upbound.io"
	MskObserverKind   = "msks.queue.entigo.com"
	KafkaTopicKind    = "topics.queue.entigo.com"
	KafkaUserKind     = "kafkausers.queue.entigo.com"
)

func testPlatformApisKafka(t *testing.T, clusterOptions *terrak8s.KubectlOptions, zonesReady <-chan struct{}, zonesReadySuccess *atomic.Bool) {
	defer func() {
		if t.Failed() {
			return
		}
		cleanupStart := time.Now()
		cleanupKafkaResources(t, clusterOptions)
		fmt.Printf("TIMING: Kafka cleanup took %s\n", time.Since(cleanupStart))
	}()

	waitForZonesReady(t, zonesReady, zonesReadySuccess)

	aAppsOptions := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, AAppsNamespace)
	namespaceOptions := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, KafkaNamespaceName)

	testWaitAndSyncKafkaApplication(t, aAppsOptions)

	// Phase 1: Wait for MSK cluster to be ready (~20-30 min)
	clusterStart := time.Now()
	waitSyncedAndReady(t, namespaceOptions, MskAwsClusterKind, KafkaClusterName, 180, 10*time.Second)
	fmt.Printf("TIMING: MSK cluster ready took %s\n", time.Since(clusterStart))
	if t.Failed() {
		return
	}

	// Read cluster ARN from status
	clusterARN := readClusterARN(t, namespaceOptions)

	// Create CRD test resources via kubectl
	createKafkaTestResources(t, clusterOptions, namespaceOptions, clusterARN)

	// Phase 2: MSK Observer tests
	observerStart := time.Now()
	runKafkaMskObserverTests(t, clusterOptions, namespaceOptions)
	fmt.Printf("TIMING: MSK observer tests took %s\n", time.Since(observerStart))
	if t.Failed() {
		return
	}

	// Phase 3: Topic + KafkaUser tests (parallel)
	topicUserStart := time.Now()
	t.Run("topic-and-user", func(t *testing.T) {
		t.Run("topic", func(t *testing.T) {
			t.Parallel()
			runKafkaTopicTests(t, namespaceOptions)
		})
		t.Run("user", func(t *testing.T) {
			t.Parallel()
			runKafkaUserTests(t, namespaceOptions)
		})
	})
	fmt.Printf("TIMING: Topic and user tests took %s\n", time.Since(topicUserStart))
}

func testWaitAndSyncKafkaApplication(t *testing.T, aAppsOptions *terrak8s.KubectlOptions) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for Application '%s' to exist", KafkaApplicationName), 30, 10*time.Second, func() (string, error) {
		name, err := terrak8s.RunKubectlAndGetOutputE(t, aAppsOptions, "get", "application", KafkaApplicationName, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if name == "" {
			return "", fmt.Errorf("application '%s' not found yet", KafkaApplicationName)
		}
		return name, nil
	})
	require.NoError(t, err, fmt.Sprintf("Application '%s' not found in namespace '%s'", KafkaApplicationName, AAppsNamespace))

	syncAndWaitApplication(t, aAppsOptions, KafkaApplicationName, 60, 10*time.Second)
}

func readClusterARN(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) string {
	t.Helper()
	arn, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", MskAwsClusterKind, KafkaClusterName, "-o", "jsonpath={.status.atProvider.arn}")
	require.NoError(t, err, "failed to read MSK cluster ARN")
	require.NotEmpty(t, arn, "MSK cluster ARN is empty")
	return arn
}

func createKafkaTestResources(t *testing.T, clusterOptions *terrak8s.KubectlOptions, namespaceOptions *terrak8s.KubectlOptions, clusterARN string) {
	t.Helper()

	// Create MSK observer (cluster-scoped)
	mskObserverYAML := fmt.Sprintf(`apiVersion: queue.entigo.com/v1alpha1
kind: MSK
metadata:
  name: %s
spec:
  clusterARN: "%s"`, MskObserverName, clusterARN)

	tmpObserverFile := fmt.Sprintf("/tmp/msk-observer-%d.yaml", time.Now().UnixNano())
	writeAndApply(t, clusterOptions, tmpObserverFile, mskObserverYAML)

	// Wait for MSK observer to be ready before creating Topic and KafkaUser
	waitSyncedAndReady(t, clusterOptions, MskObserverKind, MskObserverName, 90, 10*time.Second)

	// Create Topic (namespaced)
	topicYAML := fmt.Sprintf(`
	apiVersion: queue.entigo.com/v1alpha1
	kind: Topic
	metadata:
	name: %s
	namespace: %s
	spec:
	clusterName: %s
	partitions: 3
	replicationFactor: 2
	config:
		retention.ms: "604800000"`, KafkaTopicName, KafkaNamespaceName, KafkaClusterName)

		tmpTopicFile := fmt.Sprintf("/tmp/kafka-topic-%d.yaml", time.Now().UnixNano())
		writeAndApply(t, clusterOptions, tmpTopicFile, topicYAML)

		// Create KafkaUser (namespaced)
		kafkaUserYAML := fmt.Sprintf(`apiVersion: queue.entigo.com/v1alpha1
	kind: KafkaUser
	metadata:
	name: %s
	namespace: %s
	spec:
	clusterName: %s
	consumerGroups:
		- test-group
	acls:
		- topic: %s
		operation: Read
		- topic: %s
		operation: Write`,
	KafkaUserName, KafkaNamespaceName, KafkaClusterName, KafkaTopicName, KafkaTopicName)

	tmpUserFile := fmt.Sprintf("/tmp/kafka-user-%d.yaml", time.Now().UnixNano())
	writeAndApply(t, clusterOptions, tmpUserFile, kafkaUserYAML)
}

func writeAndApply(t *testing.T, opts *terrak8s.KubectlOptions, path string, content string) {
	t.Helper()
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err, "failed to write temp file %s", path)
	_, err = terrak8s.RunKubectlAndGetOutputE(t, opts, "apply", "-f", path)
	require.NoError(t, err, "failed to apply %s", path)
}
