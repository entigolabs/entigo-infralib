package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
)

const (
	ZoneAAppsName      = "a-apps"
	ZoneBAppsName      = "b-apps"
	ZoneAAppsNamespace = "a-apps"
	ZoneBAppsNamespace = "b-apps"
	ZoneAAppProject    = "a"
	ZoneBAppProject    = "b"
)

type appDestination struct {
	appName   string // ArgoCD Application name
	namespace string // destination namespace where workload pods run
}

var (
	zoneAChildApps = []string{"a1", "a2", "a1-default"}
	zoneBChildApps = []string{"b1", "b1-default"}

	zoneAAppDestinations = []appDestination{
		{appName: "a1", namespace: "a1"},
		{appName: "a2", namespace: "a2"},
		{appName: "a1-default", namespace: "bar"},
	}
	zoneBAppDestinations = []appDestination{
		{appName: "b1", namespace: "b1"},
		{appName: "b1-default", namespace: "bar"},
	}

	// managedNamespaceZonePool maps zone-managed namespaces to their expected
	// tenancy.entigo.com/zone-pool nodeSelector value (injected by Kyverno).
	// Pods in namespaces NOT listed here must NOT have this nodeSelector.
	managedNamespaceZonePool = map[string]string{
		"a-apps": ZoneAName + "-default",
		"b-apps": ZoneBName + "-default",
		"b1":     ZoneBName + "-default",
		"b2":     ZoneBName + "-default",
	}
)

func testZoneApps(t *testing.T, argocdNamespace string, argocdOptions *terrak8s.KubectlOptions, clusterOptions *terrak8s.KubectlOptions) {
	t.Run("a-apps", func(t *testing.T) {
		t.Parallel()
		testWaitAppProjectExists(t, argocdOptions, argocdNamespace, ZoneAAppProject)
		testApplyAndSyncZoneApp(t, argocdOptions, argocdNamespace, ZoneAAppsName, "zone_test_a_apps.yaml")
		aAppsOptions := terrak8s.NewKubectlOptions(argocdOptions.ContextName, argocdOptions.ConfigPath, ZoneAAppsNamespace)
		testWaitChildAppsReady(t, aAppsOptions, zoneAChildApps)
		testAppsHavePods(t, clusterOptions, zoneAAppDestinations)
		testPodsNodeSelector(t, clusterOptions, zoneAAppDestinations)
	})
	t.Run("b-apps", func(t *testing.T) {
		t.Parallel()
		testWaitAppProjectExists(t, argocdOptions, argocdNamespace, ZoneBAppProject)
		testApplyAndSyncZoneApp(t, argocdOptions, argocdNamespace, ZoneBAppsName, "zone_test_b_apps.yaml")
		bAppsOptions := terrak8s.NewKubectlOptions(argocdOptions.ContextName, argocdOptions.ConfigPath, ZoneBAppsNamespace)
		testWaitChildAppsReady(t, bAppsOptions, zoneBChildApps)
		testAppsHavePods(t, clusterOptions, zoneBAppDestinations)
		testPodsNodeSelector(t, clusterOptions, zoneBAppDestinations)
	})
}

func testWaitAppProjectExists(t *testing.T, opts *terrak8s.KubectlOptions, namespace, project string) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for AppProject '%s'", project), 30, 10*time.Second, func() (string, error) {
		name, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", "appproject", project, "-n", namespace, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if name == "" {
			return "", fmt.Errorf("AppProject '%s' not found", project)
		}
		return name, nil
	})
	require.NoError(t, err, fmt.Sprintf("AppProject '%s' not found", project))
}

func testApplyAndSyncZoneApp(t *testing.T, opts *terrak8s.KubectlOptions, namespace, appName, template string) {
	t.Helper()
	_, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "apply", "-f", fmt.Sprintf("./templates/%s", template), "-n", namespace)
	require.NoError(t, err, fmt.Sprintf("applying %s error", template))

	_, err = terrak8s.RunKubectlAndGetOutputE(t, opts, "patch", "application", appName, "-n", namespace, "--type", "merge", "-p", `{"operation":{"initiatedBy":{"username":"test"},"sync":{"revision":"HEAD"}}}`)
	require.NoError(t, err, fmt.Sprintf("force sync '%s' error", appName))

	_, err = retry.DoWithRetryE(t, fmt.Sprintf("waiting for Application '%s' to sync", appName), 30, 10*time.Second, func() (string, error) {
		output, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", "application", appName, "-n", namespace, "-o", "jsonpath={.status.sync.status}")
		if err != nil {
			return "", err
		}
		if output != "Synced" {
			return "", fmt.Errorf("application '%s' not synced yet, status: %s", appName, output)
		}
		return output, nil
	})
	require.NoError(t, err, fmt.Sprintf("Application '%s' failed to sync", appName))
}

func testWaitChildAppsReady(t *testing.T, opts *terrak8s.KubectlOptions, apps []string) {
	t.Helper()
	for _, app := range apps {
		waitArgoCDAppSyncedAndHealthy(t, opts, app, 30, 10*time.Second)
	}
}

func testAppsHavePods(t *testing.T, clusterOptions *terrak8s.KubectlOptions, destinations []appDestination) {
	t.Helper()
	for _, dest := range destinations {
		dest := dest
		t.Run(fmt.Sprintf("pods-%s", dest.appName), func(t *testing.T) {
			nsOpts := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, dest.namespace)
			_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for Running pods in namespace '%s'", dest.namespace), 30, 10*time.Second, func() (string, error) {
				output, err := terrak8s.RunKubectlAndGetOutputE(t, nsOpts, "get", "pods",
					"--field-selector=status.phase=Running",
					"-o", "jsonpath={.items[*].metadata.name}")
				if err != nil {
					return "", err
				}
				pods := strings.Fields(output)
				if len(pods) == 0 {
					return "", fmt.Errorf("no Running pods found in namespace '%s'", dest.namespace)
				}
				return fmt.Sprintf("found %d pod(s)", len(pods)), nil
			})
			require.NoError(t, err, fmt.Sprintf("no Running pods appeared in namespace '%s'", dest.namespace))
		})
	}
}

func testPodsNodeSelector(t *testing.T, clusterOptions *terrak8s.KubectlOptions, destinations []appDestination) {
	t.Helper()
	for _, dest := range destinations {
		dest := dest
		t.Run(fmt.Sprintf("nodeselector-%s", dest.appName), func(t *testing.T) {
			nsOpts := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, dest.namespace)
			output, err := terrak8s.RunKubectlAndGetOutputE(t, nsOpts, "get", "pods",
				"--field-selector=status.phase=Running",
				"-o", "jsonpath={.items[*].metadata.name}")
			require.NoError(t, err, fmt.Sprintf("failed to list pods in namespace '%s'", dest.namespace))

			expectedZonePool, isManaged := managedNamespaceZonePool[dest.namespace]
			for _, pod := range strings.Fields(output) {
				value, err := terrak8s.RunKubectlAndGetOutputE(t, nsOpts, "get", "pod", pod,
					"-o", `go-template={{index .spec.nodeSelector "tenancy.entigo.com/zone-pool"}}`)
				require.NoError(t, err, fmt.Sprintf("failed to get nodeSelector for pod '%s/%s'", dest.namespace, pod))
				if isManaged {
					require.Equal(t, expectedZonePool, value,
						"pod '%s/%s': nodeSelector 'tenancy.entigo.com/zone-pool' expected '%s', got '%s'",
						dest.namespace, pod, expectedZonePool, value)
				} else {
					require.Empty(t, value,
						"pod '%s/%s': nodeSelector 'tenancy.entigo.com/zone-pool' must not be set in non-managed namespace, got '%s'",
						dest.namespace, pod, value)
				}
			}
		})
	}
}

func cleanupZoneApps(t *testing.T, argocdNamespace string, argocdOptions *terrak8s.KubectlOptions) {
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocdOptions, "delete", "application", ZoneAAppsName, "-n", argocdNamespace, "--ignore-not-found")
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocdOptions, "delete", "application", ZoneBAppsName, "-n", argocdNamespace, "--ignore-not-found")
}
