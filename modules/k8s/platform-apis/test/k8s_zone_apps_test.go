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

type failingApp struct {
	name    string
	wantErr string // expected substring in operationState.message
}

var (
	zoneAChildApps = []string{"a1", "a2"}
	// a1-default destination 'bar' is not in project 'a' allowed destinations
	zoneAFailingApps = []failingApp{
		{name: "a1-default", wantErr: "not match any of the allowed destinations"},
	}
	zoneBChildApps = []string{"b1"}
	// b1-default uses source a/a1 whose resources target namespace a1 (zone A) — zone B project blocks it
	zoneBFailingApps = []failingApp{
		{name: "b1-default", wantErr: "not valid"},
	}

	zoneAAppDestinations = []appDestination{
		{appName: "a1", namespace: "a1"},
		{appName: "a2", namespace: "a2"},
	}
	zoneBAppDestinations = []appDestination{
		{appName: "b1", namespace: "b1"},
	}

	// managedNamespaceZonePool maps zone-managed namespaces to their expected
	// tenancy.entigo.com/zone-pool nodeSelector value (injected by Kyverno).
	// Pods in namespaces NOT listed here must NOT have this nodeSelector.
	managedNamespaceZonePool = map[string]string{
		"a-apps": ZoneAName + "-default",
		"a1":     ZoneAName + "-default",
		"a2":     ZoneAName + "-default",
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
		for _, app := range zoneAFailingApps {
			testWaitChildAppSyncFailed(t, aAppsOptions, app.name, app.wantErr)
		}
		testAppsHavePods(t, clusterOptions, zoneAAppDestinations)
		testPodsNodeSelector(t, clusterOptions, zoneAAppDestinations)
	})
	t.Run("b-apps", func(t *testing.T) {
		t.Parallel()
		testWaitAppProjectExists(t, argocdOptions, argocdNamespace, ZoneBAppProject)
		testApplyAndSyncZoneApp(t, argocdOptions, argocdNamespace, ZoneBAppsName, "zone_test_b_apps.yaml")
		bAppsOptions := terrak8s.NewKubectlOptions(argocdOptions.ContextName, argocdOptions.ConfigPath, ZoneBAppsNamespace)
		testWaitChildAppsReady(t, bAppsOptions, zoneBChildApps)
		for _, app := range zoneBFailingApps {
			testWaitChildAppSyncFailed(t, bAppsOptions, app.name, app.wantErr)
		}
		testAppsHavePods(t, clusterOptions, zoneBAppDestinations)
		testPodsNodeSelector(t, clusterOptions, zoneBAppDestinations)
	})
	t.Run("nodeselector-enforcement", func(t *testing.T) {
		t.Parallel()
		testWrongNodeSelectorRejected(t, clusterOptions)
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
		forceSyncArgoApp(t, opts, app)
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

// testWaitChildAppSyncFailed force-syncs an app and verifies it enters an error state
// whose message contains wantErr. It handles two ArgoCD failure modes:
//   - operationState.phase == "Failed": sync ran but was rejected (e.g. b1-default trying
//     to deploy resources into a zone A namespace not permitted in project b)
//   - status.conditions[InvalidSpecError]: spec rejected before sync starts (e.g. a1-default
//     whose destination 'bar' is not in project a's allowed destinations)
func testWaitChildAppSyncFailed(t *testing.T, opts *terrak8s.KubectlOptions, appName, wantErr string) {
	t.Helper()
	forceSyncArgoApp(t, opts, appName)
	var errMsg string
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for app '%s' to enter error state", appName), 30, 10*time.Second, func() (string, error) {
		// Check for sync operation failure
		phase, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", "application", appName,
			"-o", "jsonpath={.status.operationState.phase}")
		if err != nil {
			return "", err
		}
		if phase == "Failed" {
			msg, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", "application", appName,
				"-o", "jsonpath={.status.operationState.message}")
			if err != nil {
				return "", err
			}
			errMsg = msg
			return phase, nil
		}
		// Check for spec-level validation error stored in conditions
		cond, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", "application", appName,
			"-o", `jsonpath={.status.conditions[?(@.type=="InvalidSpecError")].message}`)
		if err != nil {
			return "", err
		}
		if cond != "" {
			errMsg = cond
			return "InvalidSpecError", nil
		}
		return "", fmt.Errorf("app '%s' not in error state yet (phase=%q)", appName, phase)
	})
	require.NoError(t, err, fmt.Sprintf("app '%s' was expected to enter an error state", appName))
	require.Contains(t, errMsg, wantErr,
		"app '%s' error message does not match expected; got: %s", appName, errMsg)
}

func testWrongNodeSelectorRejected(t *testing.T, clusterOptions *terrak8s.KubectlOptions) {
	t.Helper()
	nsOpts := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, ZoneAAppsNamespace)
	defer terrak8s.RunKubectl(t, nsOpts, "delete", "pod", "test-wrong-nodeselector", "--ignore-not-found")

	_, err := terrak8s.RunKubectlAndGetOutputE(t, nsOpts, "apply", "-f", "./templates/zone_test_wrong_nodeselector_pod.yaml")
	require.Error(t, err, "pod with zone B nodeSelector in zone A namespace should be rejected by Kyverno")
}

func cleanupZoneApps(t *testing.T, argocdNamespace string, argocdOptions *terrak8s.KubectlOptions) {
	aAppsOpts := terrak8s.NewKubectlOptions(argocdOptions.ContextName, argocdOptions.ConfigPath, ZoneAAppsNamespace)
	for _, app := range zoneAChildApps {
		_, _ = terrak8s.RunKubectlAndGetOutputE(t, aAppsOpts, "delete", "application", app, "--ignore-not-found")
	}
	for _, fa := range zoneAFailingApps {
		_, _ = terrak8s.RunKubectlAndGetOutputE(t, aAppsOpts, "delete", "application", fa.name, "--ignore-not-found")
	}
	bAppsOpts := terrak8s.NewKubectlOptions(argocdOptions.ContextName, argocdOptions.ConfigPath, ZoneBAppsNamespace)
	for _, app := range zoneBChildApps {
		_, _ = terrak8s.RunKubectlAndGetOutputE(t, bAppsOpts, "delete", "application", app, "--ignore-not-found")
	}
	for _, fa := range zoneBFailingApps {
		_, _ = terrak8s.RunKubectlAndGetOutputE(t, bAppsOpts, "delete", "application", fa.name, "--ignore-not-found")
	}
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocdOptions, "delete", "application", ZoneAAppsName, "-n", argocdNamespace, "--ignore-not-found")
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocdOptions, "delete", "application", ZoneBAppsName, "-n", argocdNamespace, "--ignore-not-found")
}
