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
	ClusterProviderConfigKind = "clusterproviderconfig.kafka.m.crossplane.io"
	KafkaNs                   = "crossplane-kafka"
)

func runKafkaMskObserverTests(t *testing.T, clusterOptions *terrak8s.KubectlOptions, namespaceOptions *terrak8s.KubectlOptions) {
	kafkaNsOptions := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, KafkaNs)

	t.Run("msk-observer", func(t *testing.T) {
		t.Run("sub-resources", func(t *testing.T) {
			t.Run("observed-cluster", func(t *testing.T) {
				t.Parallel()
				testMskObservedClusterSyncedAndReady(t, clusterOptions)
			})
			t.Run("config-secret", func(t *testing.T) {
				t.Parallel()
				testMskConfigSecretExists(t, kafkaNsOptions)
			})
			t.Run("cluster-provider-config", func(t *testing.T) {
				t.Parallel()
				testMskClusterProviderConfigExists(t, clusterOptions)
			})
		})

		if t.Failed() {
			return
		}

		testMskStatusFields(t, clusterOptions)
	})
}

func testMskObservedClusterSyncedAndReady(t *testing.T, clusterOptions *terrak8s.KubectlOptions) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for observed Cluster (composite=%s)", MskObserverName), 60, 10*time.Second, func() (string, error) {
		name, err := getFirstByLabel(t, clusterOptions, MskAwsClusterKind, MskObserverName)
		if err != nil {
			return "", err
		}
		if name == "" {
			return "", fmt.Errorf("no observed Cluster found for composite=%s", MskObserverName)
		}
		for _, condType := range []string{"Synced", "Ready"} {
			status, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", MskAwsClusterKind, name, "-o",
				fmt.Sprintf(`jsonpath={.status.conditions[?(@.type=="%s")].status}`, condType))
			if err != nil {
				return "", err
			}
			if status != "True" {
				return "", fmt.Errorf("observed Cluster '%s': %s=%s", name, condType, status)
			}
		}
		return name, nil
	})
	require.NoError(t, err, "observed Cluster for MSK observer '%s' failed to become Synced and Ready", MskObserverName)
}

func testMskConfigSecretExists(t *testing.T, kafkaNsOptions *terrak8s.KubectlOptions) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for config Secret (composite=%s)", MskObserverName), 60, 10*time.Second, func() (string, error) {
		name, err := getFirstByLabel(t, kafkaNsOptions, "secret", MskObserverName)
		if err != nil {
			return "", err
		}
		if name == "" {
			return "", fmt.Errorf("no config Secret found for composite=%s in namespace=%s", MskObserverName, KafkaNs)
		}
		return name, nil
	})
	require.NoError(t, err, "config Secret for MSK observer '%s' not found in namespace '%s'", MskObserverName, KafkaNs)
}

func testMskClusterProviderConfigExists(t *testing.T, clusterOptions *terrak8s.KubectlOptions) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for ClusterProviderConfig (composite=%s)", MskObserverName), 60, 10*time.Second, func() (string, error) {
		name, err := getFirstByLabel(t, clusterOptions, ClusterProviderConfigKind, MskObserverName)
		if err != nil {
			return "", err
		}
		if name == "" {
			return "", fmt.Errorf("no ClusterProviderConfig found for composite=%s", MskObserverName)
		}
		return name, nil
	})
	require.NoError(t, err, "ClusterProviderConfig for MSK observer '%s' not found", MskObserverName)
}

func testMskStatusFields(t *testing.T, clusterOptions *terrak8s.KubectlOptions) {
	t.Helper()

	fields := []struct {
		jsonPath string
		label    string
	}{
		{"jsonpath={.status.brokers}", "brokers"},
		{"jsonpath={.status.brokersscram}", "brokersscram"},
		{"jsonpath={.status.arn}", "arn"},
		{"jsonpath={.status.region}", "region"},
		{"jsonpath={.status.providerConfig}", "providerConfig"},
	}

	for _, field := range fields {
		value, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", MskObserverKind, MskObserverName, "-o", field.jsonPath)
		require.NoError(t, err, "failed to get MSK observer status.%s", field.label)
		require.NotEmpty(t, value, "MSK observer '%s' status.%s is empty", MskObserverName, field.label)
	}
}
