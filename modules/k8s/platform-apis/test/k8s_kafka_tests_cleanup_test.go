package test

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
)

func cleanupKafkaResources(t *testing.T, clusterOptions *terrak8s.KubectlOptions) {
	kafkaNsOptions := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, KafkaNamespaceName)

	// Phase 1: Delete KafkaUser
	cleanupKafkaDeleteForeground(t, kafkaNsOptions, KafkaUserKind, KafkaUserName)
	cleanupKafkaWaitForDeletion(t, kafkaNsOptions, KafkaUserKind, KafkaUserName, 60)
	cleanupWaitForKafkaUserSubResources(t, kafkaNsOptions)

	// Phase 2: Delete Topic
	cleanupKafkaDeleteForeground(t, kafkaNsOptions, KafkaTopicKind, KafkaTopicName)
	cleanupKafkaWaitForDeletion(t, kafkaNsOptions, KafkaTopicKind, KafkaTopicName, 30)

	// Phase 3: Delete MSK observer (cluster-scoped)
	cleanupKafkaDeleteForeground(t, clusterOptions, MskObserverKind, MskObserverName)
	cleanupKafkaWaitForDeletion(t, clusterOptions, MskObserverKind, MskObserverName, 60)
	cleanupWaitForMskObserverSubResources(t, clusterOptions)

	// Phase 4: Delete the real MSK cluster
	cleanupKafkaDeleteForeground(t, kafkaNsOptions, MskAwsClusterKind, KafkaClusterName)
	cleanupKafkaWaitForDeletion(t, kafkaNsOptions, MskAwsClusterKind, KafkaClusterName, 180)

	// Phase 5: Verify all managed resources gone
	if !cleanupKafkaFullyGone(t, kafkaNsOptions, clusterOptions) {
		fmt.Printf("WARNING: some Kafka managed resources still exist; skipping namespace deletion to avoid finalizer deadlock\n")
		return
	}

	// Phase 6: Namespace and Application cleanup
	cleanupKafkaNamespace(t, kafkaNsOptions, clusterOptions)

	aAppsOptions := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, AAppsNamespace)
	fmt.Printf("Cleanup: deleting Kafka Application '%s' from '%s'\n", KafkaApplicationName, AAppsNamespace)
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, aAppsOptions, "delete", "application", KafkaApplicationName, "--ignore-not-found")
}

func cleanupKafkaDeleteForeground(t *testing.T, opts *terrak8s.KubectlOptions, kind string, name string) {
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, opts, "delete", kind, name, "--cascade=foreground", "--wait=false", "--ignore-not-found")
}

func cleanupKafkaWaitForDeletion(t *testing.T, opts *terrak8s.KubectlOptions, kind string, name string, maxRetries int) {
	_, _ = retry.DoWithRetryE(t, fmt.Sprintf("waiting for %s/%s deletion", kind, name), maxRetries, 10*time.Second, func() (string, error) {
		output, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", kind, name, "--ignore-not-found", "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if output != "" {
			return "", fmt.Errorf("%s/%s still exists", kind, name)
		}
		return "deleted", nil
	})
}

func cleanupWaitForKafkaUserSubResources(t *testing.T, opts *terrak8s.KubectlOptions) {
	subResourceKinds := []struct {
		kind  string
		label string
	}{
		{AwsSecretKind, "AWS Secrets"},
		{SecretVersionKind, "Secret Versions"},
		{SecretPolicyKind, "Secret Policies"},
		{ScramAssocKind, "SCRAM Associations"},
		{AclKind, "ACLs"},
	}

	var wg sync.WaitGroup
	for _, srk := range subResourceKinds {
		srk := srk
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = retry.DoWithRetryE(t, fmt.Sprintf("waiting for %s deletion", srk.label), 60, 10*time.Second, func() (string, error) {
				output, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", srk.kind, "-l",
					fmt.Sprintf("crossplane.io/composite=%s", KafkaUserName), "-o", "jsonpath={.items[*].metadata.name}", "--ignore-not-found")
				if err != nil {
					return "", err
				}
				if output != "" {
					return "", fmt.Errorf("%s still exist: %s", srk.label, output)
				}
				return "deleted", nil
			})
		}()
	}
	wg.Wait()

	// Also wait for the k8s secret
	secretName := fmt.Sprintf("%s-%s-k8s", KafkaClusterName, KafkaUserName)
	_, _ = retry.DoWithRetryE(t, fmt.Sprintf("waiting for k8s Secret '%s' deletion", secretName), 30, 10*time.Second, func() (string, error) {
		output, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", "secret", secretName, "--ignore-not-found", "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if output != "" {
			return "", fmt.Errorf("k8s Secret '%s' still exists", secretName)
		}
		return "deleted", nil
	})
}

func cleanupWaitForMskObserverSubResources(t *testing.T, clusterOptions *terrak8s.KubectlOptions) {
	subResourceKinds := []struct {
		kind  string
		label string
	}{
		{MskAwsClusterKind, "observed Clusters"},
		{ClusterProviderConfigKind, "ClusterProviderConfigs"},
	}

	var wg sync.WaitGroup
	for _, srk := range subResourceKinds {
		srk := srk
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = retry.DoWithRetryE(t, fmt.Sprintf("waiting for %s deletion", srk.label), 60, 10*time.Second, func() (string, error) {
				output, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", srk.kind, "-l",
					fmt.Sprintf("crossplane.io/composite=%s", MskObserverName), "-o", "jsonpath={.items[*].metadata.name}", "--ignore-not-found")
				if err != nil {
					return "", err
				}
				if output != "" {
					return "", fmt.Errorf("%s still exist: %s", srk.label, output)
				}
				return "deleted", nil
			})
		}()
	}
	wg.Wait()

	// Also wait for config secret in kafka namespace
	kafkaNsOptions := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, KafkaNs)
	_, _ = retry.DoWithRetryE(t, "waiting for config Secret deletion", 30, 10*time.Second, func() (string, error) {
		output, err := terrak8s.RunKubectlAndGetOutputE(t, kafkaNsOptions, "get", "secret", "-l",
			fmt.Sprintf("crossplane.io/composite=%s", MskObserverName), "-o", "jsonpath={.items[*].metadata.name}", "--ignore-not-found")
		if err != nil {
			return "", err
		}
		if output != "" {
			return "", fmt.Errorf("config Secrets still exist: %s", output)
		}
		return "deleted", nil
	})
}

func cleanupKafkaFullyGone(t *testing.T, kafkaNsOptions *terrak8s.KubectlOptions, clusterOptions *terrak8s.KubectlOptions) bool {
	// Check namespaced resources
	namespacedKinds := []string{MskAwsClusterKind, AwsSecretKind, SecretVersionKind, SecretPolicyKind, ScramAssocKind, AclKind, KafkaTopicMRKind}
	for _, kind := range namespacedKinds {
		out, err := terrak8s.RunKubectlAndGetOutputE(t, kafkaNsOptions, "get", kind, "-o", "jsonpath={.items[*].metadata.name}", "--ignore-not-found")
		if err != nil || out != "" {
			return false
		}
	}

	// Check cluster-scoped resources
	clusterScopedComposites := []string{MskObserverName}
	for _, composite := range clusterScopedComposites {
		out, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", MskObserverKind, composite, "--ignore-not-found", "-o", "jsonpath={.metadata.name}")
		if err != nil || out != "" {
			return false
		}
	}

	return true
}

func cleanupKafkaNamespace(t *testing.T, kafkaNsOptions *terrak8s.KubectlOptions, clusterOptions *terrak8s.KubectlOptions) {
	leftovers, _ := terrak8s.RunKubectlAndGetOutputE(t, kafkaNsOptions, "get", "all", "-n", KafkaNamespaceName, "--ignore-not-found", "-o", "name")
	if leftovers != "" {
		names := strings.Fields(leftovers)
		fmt.Printf("Cleanup: deleting %d leftover resources in namespace '%s'\n", len(names), KafkaNamespaceName)
		_, _ = terrak8s.RunKubectlAndGetOutputE(t, kafkaNsOptions, "delete", "all", "--all", "-n", KafkaNamespaceName, "--cascade=foreground", "--wait=false", "--ignore-not-found")
		time.Sleep(10 * time.Second)
	}
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "delete", "namespace", KafkaNamespaceName, "--ignore-not-found", "--wait=true")
}
