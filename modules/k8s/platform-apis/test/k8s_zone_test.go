package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
)

func testZone(t *testing.T, ctx context.Context, cluster *terrak8s.KubectlOptions) {
	if ctx.Err() != nil {
		return
	}
	testZoneApps(t, cluster)
	if t.Failed() {
		return
	}
	testZoneKyverno(t, cluster)
}

func testZoneApps(t *testing.T, cluster *terrak8s.KubectlOptions) {
	aApps := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, AAppsNamespace)
	bApps := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, BAppsNamespace)

	deployAndVerifyApp(t, cluster, aApps, "./templates/a_test_application.yaml", AAppsApplicationName)
	deployAndVerifyApp(t, cluster, bApps, "./templates/b_test_application.yaml", BAppsApplicationName)

	verifyAppsRunning(t, cluster)
	verifyNetworkPolicies(t, cluster)
}

func testPodsRunning(t *testing.T, cluster *terrak8s.KubectlOptions, namespace, podName string) {
	t.Helper()
	nsOpts := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, namespace)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("pod %s/%s Running", namespace, podName), 10, 10*time.Second,
		func() (string, error) {
			phase, err := terrak8s.RunKubectlAndGetOutputE(t, nsOpts, "get", "pod", podName, "-o", "jsonpath={.status.phase}")
			if err != nil {
				return "", err
			}
			if phase != "Running" {
				return "", fmt.Errorf("phase=%q", phase)
			}
			return phase, nil
		})
	require.NoError(t, err, "pod %s/%s never reached Running", namespace, podName)
}

func deployAndVerifyApp(t *testing.T, cluster, appOpts *terrak8s.KubectlOptions, templatePath, appName string) {
	t.Helper()
	applyFile(t, cluster, templatePath)
	syncWithRetry(t, appOpts, appName)
	waitApplicationHealthy(t, appOpts, appName)
}

func verifyAppsRunning(t *testing.T, cluster *terrak8s.KubectlOptions) {
	t.Run("apps-running", func(t *testing.T) {
		t.Run("a1", func(t *testing.T) {
			t.Parallel()
			testPodsRunning(t, cluster, "a1", "a1-curl")
		})
		t.Run("b1", func(t *testing.T) {
			t.Parallel()
			testPodsRunning(t, cluster, "b1", "b1-curl")
		})
	})
}

func verifyNetworkPolicies(t *testing.T, cluster *terrak8s.KubectlOptions) {
	t.Run("network-policies", func(t *testing.T) {
		t.Run("a1", func(t *testing.T) {
			t.Parallel()
			testNetworkPolicyMatchLabels(t, cluster, "a1", map[string]string{"tenancy.entigo.com/zone": "a"})
		})
		t.Run("b1", func(t *testing.T) {
			t.Parallel()
			testNetworkPolicyMatchLabels(t, cluster, "b1", map[string]string{"kubernetes.io/metadata.name": "b1"})
		})
	})
}

func testNetworkPolicyMatchLabels(t *testing.T, cluster *terrak8s.KubectlOptions, namespace string, expectedLabels map[string]string) {
	t.Helper()
	nsOpts := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, namespace)
	name := fmt.Sprintf("%s-zone", namespace)

	policy, err := terrak8s.GetNetworkPolicyE(t, nsOpts, name)
	require.NoError(t, err, "Failed to get network policy")

	require.NotEmpty(t, policy.Spec.Ingress, "Ingress rules should not be empty")
	require.NotEmpty(t, policy.Spec.Ingress[0].From, "From peers should not be empty")
	require.NotNil(t, policy.Spec.Ingress[0].From[0].NamespaceSelector, "NamespaceSelector should not be nil")

	actualLabels := policy.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels
	require.Equal(t, expectedLabels, actualLabels, "NetworkPolicy matchLabels do not match expected values")
}
