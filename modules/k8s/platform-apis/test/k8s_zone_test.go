package test

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
)

type testZones struct {
	prefix, domain, gateway, gatewayNamespace, serviceIngressClass string
	secretStoreWrite                                               bool
}

type zoneApplicationData struct {
	Namespace, Prefix string
}

func (z testZones) Prefix() string {
	return z.prefix
}

func (z testZones) Domain() string {
	return z.domain
}

func (z testZones) Gateway() string {
	return z.gateway
}

func (z testZones) GatewayNamespace() string {
	return z.gatewayNamespace
}

func (z testZones) ServiceIngressClass() string {
	return z.serviceIngressClass
}

func (z testZones) A() string {
	return fmt.Sprintf("%s%s", z.prefix, ZoneAName)
}

func (z testZones) B() string {
	return fmt.Sprintf("%s%s", z.prefix, ZoneBName)
}

func (z testZones) AApps() string {
	return fmt.Sprintf("%s-apps", z.A())
}

func (z testZones) BApps() string {
	return fmt.Sprintf("%s-apps", z.B())
}

func (z testZones) PoolA() string {
	return fmt.Sprintf("%s-%s", z.A(), ZoneAPool)
}

func (z testZones) PoolB() string {
	return fmt.Sprintf("%s-%s", z.B(), ZoneBPool)
}

func (z testZones) SecretStoreA() string {
	return fmt.Sprintf("zone-%s", z.A())
}

func (z testZones) RemoteKeyA() string {
	return fmt.Sprintf("%s%s", SecretStoreRemoteKeyPrefix, z.A())
}

func (z testZones) RemoteKeyB() string {
	return fmt.Sprintf("%s%s", SecretStoreRemoteKeyPrefix, z.B())
}

func (z testZones) RemoteKeyUntagged() string {
	return fmt.Sprintf("%s%suntagged", SecretStoreRemoteKeyPrefix, z.prefix)
}

func testZone(t *testing.T, ctx context.Context, cluster, argocd *terrak8s.KubectlOptions, z testZones) {
	if ctx.Err() != nil {
		return
	}
	testZoneApps(t, cluster, z)
	if t.Failed() {
		return
	}
	testZoneKyverno(t, cluster, z)
	testZoneRBAC(t, cluster, argocd, z)
	if t.Failed() {
		return
	}
	if !z.secretStoreWrite {
		t.Run("secret-store", func(t *testing.T) {
			t.Skip("the cluster-wide store of this environment is read-only, the fixtures cannot be pushed")
		})
		return
	}
	testZoneSecretStore(t, cluster, z)
}

func testZoneApps(t *testing.T, cluster *terrak8s.KubectlOptions, z testZones) {
	aApps := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, z.AApps())
	bApps := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, z.BApps())

	deployAndVerifyApp(t, cluster, aApps, writeTempYAML(t, renderTemplate(t, "./templates/a_test_application.yaml", z)), AAppsApplicationName)
	deployAndVerifyApp(t, cluster, bApps, writeTempYAML(t, renderTemplate(t, "./templates/b_test_application.yaml", z)), BAppsApplicationName)

	verifySourceReposAllowList(t, cluster, z)
	verifyAppsRunning(t, cluster, z)
	verifyNetworkPolicies(t, cluster, z)
}

func testPodsRunning(t *testing.T, nsOpts *terrak8s.KubectlOptions, podName string) {
	t.Helper()
	// 30 attempts, like the other waits in this suite. The Pod still has to be created by its own
	// child Application, pull the app and istio-proxy images onto a node that may have just joined,
	// and get its sidecar through the startup probe.
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("pod %s/%s Running", nsOpts.Namespace, podName), 30, 10*time.Second,
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
	require.NoError(t, err, "pod %s/%s never reached Running", nsOpts.Namespace, podName)
}

func deployAndVerifyApp(t *testing.T, cluster, appOpts *terrak8s.KubectlOptions, templatePath, appName string) {
	t.Helper()
	applyFile(t, cluster, templatePath)
	syncWithRetry(t, appOpts, appName)
	waitApplicationHealthy(t, appOpts, appName)
}

// verifySourceReposAllowList checks the environment-wide AppProject sourceRepos allow-list
// (global.sourceRepos, set to ghcr.io/entigolabs for these tests). The allowed-repo case is
// already covered above: every test app (repoURL ghcr.io/entigolabs) only syncs because that
// repo is in the list. Here we assert the negative: an Application pointing at a repo outside
// the list is rejected by ArgoCD and never deploys.
func verifySourceReposAllowList(t *testing.T, cluster *terrak8s.KubectlOptions, z testZones) {
	t.Run("sourcerepos-allow-list", func(t *testing.T) {
		bApps := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, z.BApps())
		applyFile(t, cluster, writeTempYAML(t, renderTemplate(t, "./templates/b_denied_application.yaml", z)))
		defer func() {
			_, _ = terrak8s.RunKubectlAndGetOutputE(t, bApps, "delete", "application", BDeniedApplicationName, "--ignore-not-found")
		}()
		_, err := retry.DoWithRetryE(t, fmt.Sprintf("Application %q rejected by project", BDeniedApplicationName),
			healthRetries, healthInterval, func() (string, error) {
				msg, err := getAppField(t, bApps, BDeniedApplicationName, ".status.conditions[*].message")
				if err != nil {
					return "", err
				}
				if !strings.Contains(msg, "is not permitted in project") {
					return "", fmt.Errorf("conditions=%q", msg)
				}
				return msg, nil
			})
		require.NoError(t, err, "Application %q using a repo outside the allow-list was not rejected", BDeniedApplicationName)
	})
}

func verifyAppsRunning(t *testing.T, cluster *terrak8s.KubectlOptions, z testZones) {
	t.Run("apps-running", func(t *testing.T) {
		t.Run("a1", func(t *testing.T) {
			t.Parallel()
			nsOpts := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, "a1")
			testPodsRunning(t, nsOpts, "a1-curl")
			testPodNodeSelector(t, nsOpts, "a1-curl", "tenancy.entigo.com/zone", z.A())
		})
		t.Run("b1", func(t *testing.T) {
			t.Parallel()
			nsOpts := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, "b1")
			testPodsRunning(t, nsOpts, "b1-curl")
			testPodNodeSelector(t, nsOpts, "b1-curl", "tenancy.entigo.com/zone-pool", z.PoolB())
		})
	})
}

func testPodNodeSelector(t *testing.T, nsOpts *terrak8s.KubectlOptions, podName, key, value string) {
	t.Helper()
	pod, err := terrak8s.GetPodE(t, nsOpts, podName)
	require.NoError(t, err, "Failed to get pod %s/%s", nsOpts.Namespace, podName)
	require.Equal(t, value, pod.Spec.NodeSelector[key],
		"pod %s/%s nodeSelector[%q] mismatch (full: %v)", nsOpts.Namespace, podName, key, pod.Spec.NodeSelector)
}

func verifyNetworkPolicies(t *testing.T, cluster *terrak8s.KubectlOptions, z testZones) {
	t.Run("network-policies", func(t *testing.T) {
		t.Run("a1", func(t *testing.T) {
			t.Parallel()
			testNetworkPolicyMatchLabels(t, cluster, "a1", map[string]string{"tenancy.entigo.com/zone": z.A()})
		})
		t.Run("b1", func(t *testing.T) {
			t.Parallel()
			testNetworkPolicyMatchLabels(t, cluster, "b1", map[string]string{"kubernetes.io/metadata.name": "b1"})
		})
		t.Run("a1-target", func(t *testing.T) {
			t.Parallel()
			testTargetNetworkPolicy(t, cluster, "a1", "a1-a1-8080", "a1", "8080")
			testTargetNetworkPolicy(t, cluster, "a1", "a1-a1second-http", "a1second", "http")
		})
		t.Run("b1-target", func(t *testing.T) {
			t.Parallel()
			testTargetNetworkPolicy(t, cluster, "b1", "b1-b1-8080", "b1", "8080")
			testTargetNetworkPolicy(t, cluster, "b1", "b1-b1second-http", "b1second", "http")
		})
		t.Run("a1-route-target", func(t *testing.T) {
			t.Parallel()
			testTargetNetworkPolicy(t, cluster, "a1", "a1-route-a1-8080", "a1", "8080")
			testTargetNetworkPolicy(t, cluster, "a1", "a1-route-a1second-http", "a1second", "http")
		})
		t.Run("b1-route-target", func(t *testing.T) {
			t.Parallel()
			testTargetNetworkPolicy(t, cluster, "b1", "b1-route-b1-8080", "b1", "8080")
			testTargetNetworkPolicy(t, cluster, "b1", "b1-route-b1second-http", "b1second", "http")
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

// testTargetNetworkPolicy verifies the NetworkPolicy generated for an Ingress path or an HTTPRoute
// backend, named <ingress|httpRoute>-<service>-<targetPort>. Its source peers come from the load
// balancer subnets (IngressClass -> IngressClassParams -> Subnet, or Gateway ->
// LoadBalancerConfiguration -> Subnet), which differ per environment, so only the shape of the peers
// is asserted here.
func testTargetNetworkPolicy(t *testing.T, cluster *terrak8s.KubectlOptions, namespace, name, expectedApp, expectedPort string) {
	t.Helper()
	nsOpts := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, namespace)
	// Generated only after the tenancy function observes the Ingress and Service
	waitResourceExists(t, nsOpts, "networkpolicy", name, healthRetries, healthInterval)

	policy, err := terrak8s.GetNetworkPolicyE(t, nsOpts, name)
	require.NoError(t, err, "Failed to get target network policy %s/%s", namespace, name)

	require.Equal(t, map[string]string{"app": expectedApp}, policy.Spec.PodSelector.MatchLabels,
		"podSelector mismatch for %s/%s", namespace, name)
	require.Len(t, policy.Spec.Ingress, 1, "expected a single ingress rule in %s/%s", namespace, name)
	require.Len(t, policy.Spec.Ingress[0].Ports, 1, "expected a single port in %s/%s", namespace, name)
	require.NotNil(t, policy.Spec.Ingress[0].Ports[0].Port)
	require.Equal(t, expectedPort, policy.Spec.Ingress[0].Ports[0].Port.String(),
		"port mismatch for %s/%s", namespace, name)
	require.NotNil(t, policy.Spec.Ingress[0].Ports[0].Protocol)
	require.Equal(t, "TCP", string(*policy.Spec.Ingress[0].Ports[0].Protocol))

	// An empty From means allow from anywhere
	require.NotEmpty(t, policy.Spec.Ingress[0].From, "%s/%s must restrict sources to the load balancer subnets", namespace, name)
	for _, peer := range policy.Spec.Ingress[0].From {
		require.NotNil(t, peer.IPBlock, "%s/%s peers must be ipBlocks", namespace, name)
		_, _, err := net.ParseCIDR(peer.IPBlock.CIDR)
		require.NoError(t, err, "invalid CIDR %q in %s/%s", peer.IPBlock.CIDR, namespace, name)
		require.Nil(t, peer.NamespaceSelector, "%s/%s peers must not select namespaces", namespace, name)
		require.Nil(t, peer.PodSelector, "%s/%s peers must not select pods", namespace, name)
	}
}
