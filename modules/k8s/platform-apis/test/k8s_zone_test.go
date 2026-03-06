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
	AppProjectName        = "zone"
	ZoneApplicationName   = "app-of-zones"
	ZoneAName             = "a"
	ZoneBName             = "b"
	ZoneConfigurationName = "platform-apis-zone"
	ZoneKind              = "zone.tenancy.entigo.com"
	TenancyFunctionName   = "platform-apis-tenancy-fn"
	FunctionKind          = "function.pkg.crossplane.io"

	AAppsNamespace       = "a-apps"
	BAppsNamespace       = "b-apps"
	AAppsApplicationName = "a-apps"
	BAppsApplicationName = "b-apps"
)

func testPlatformApisZone(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions, argocdOptions *terrak8s.KubectlOptions, signalZonesReady func(bool)) {
	//defer cleanupZoneResources(t, argocdNamespace, argocdOptions, clusterOptions)

	_, err := terrak8s.RunKubectlAndGetOutputE(t, argocdOptions, "get", "appproject", AppProjectName, "-n", argocdNamespace)
	require.NoError(t, err, "AppProject '%s' not found in ArgoCD namespace '%s'", AppProjectName, argocdNamespace)

	_, err = terrak8s.RunKubectlAndGetOutputE(t, argocdOptions, "apply", "-f", "./templates/zone_test_application.yaml")
	require.NoError(t, err, "applying zone Application error")

	syncAndWaitApplication(t, argocdOptions, ZoneApplicationName, 30, 10*time.Second)

	// Wait for both zones to be ready in parallel; t.Run blocks until all parallel subtests finish
	t.Run("zones-ready", func(t *testing.T) {
		for _, zone := range []string{ZoneAName, ZoneBName} {
			zone := zone
			t.Run(zone, func(t *testing.T) {
				t.Parallel()
				waitSyncedAndReady(t, clusterOptions, ZoneKind, zone, 30, 10*time.Second)
				testZoneNodegroupReady(t, clusterOptions, zone)
			})
		}
	})

	if t.Failed() {
		signalZonesReady(false)
		return
	}

	aAppsOpts := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, AAppsNamespace)
	bAppsOpts := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, BAppsNamespace)

	_, err = terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "apply", "-f", "./templates/a_test_application.yaml")
	require.NoError(t, err, "applying a-apps Application error")

	_, err = terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "apply", "-f", "./templates/b_test_application.yaml")
	require.NoError(t, err, "applying b-apps Application error")

	// Sync a-apps first — its child 'test-postgresql' Application must exist before postgresql tests start
	syncAndWaitApplication(t, aAppsOpts, AAppsApplicationName, 30, 10*time.Second)
	signalZonesReady(!t.Failed())

	if t.Failed() {
		return
	}

	syncAndWaitApplication(t, bAppsOpts, BAppsApplicationName, 30, 10*time.Second)

	t.Run("zone-apps-running", func(t *testing.T) {
		t.Run("a1", func(t *testing.T) {
			t.Parallel()
			syncAndWaitApplication(t, aAppsOpts, "a1", 30, 10*time.Second)
			testPodsRunning(t, clusterOptions, "a1", "a1")
		})
		t.Run("b1", func(t *testing.T) {
			t.Parallel()
			syncAndWaitApplication(t, bAppsOpts, "b1", 30, 10*time.Second)
			testPodsRunning(t, clusterOptions, "b1", "b1")
		})
	})
}

func testZoneNodegroupReady(t *testing.T, clusterOptions *terrak8s.KubectlOptions, zone string) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for zone '%s' NodeGroup", zone), 30, 10*time.Second, func() (string, error) {
		names, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", "nodegroup.eks.aws.upbound.io",
			"-l", fmt.Sprintf("crossplane.io/composite=%s", zone), "-o", "jsonpath={.items[*].metadata.name}")
		if err != nil {
			return "", err
		}
		if names == "" {
			return "", fmt.Errorf("zone '%s' has no NodeGroups", zone)
		}
		status, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", "nodegroup.eks.aws.upbound.io",
			"-l", fmt.Sprintf("crossplane.io/composite=%s", zone), "-o", `jsonpath={.items[0].status.conditions[?(@.type=="Ready")].status}`)
		if err != nil {
			return "", err
		}
		if status != "True" {
			return "", fmt.Errorf("zone '%s' NodeGroup not ready: %s", zone, status)
		}
		return status, nil
	})
	require.NoError(t, err, "zone '%s' NodeGroup not ready", zone)
}

func testPodsRunning(t *testing.T, clusterOptions *terrak8s.KubectlOptions, namespace, releaseName string) {
	t.Helper()
	nsOpts := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, namespace)
	for _, pod := range []string{releaseName, releaseName + "second"} {
		_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for pod '%s/%s'", namespace, pod), 30, 10*time.Second, func() (string, error) {
			phase, err := terrak8s.RunKubectlAndGetOutputE(t, nsOpts, "get", "pod", pod, "-o", "jsonpath={.status.phase}")
			if err != nil {
				return "", err
			}
			if phase != "Running" {
				return "", fmt.Errorf("pod phase=%s", phase)
			}
			return phase, nil
		})
		require.NoError(t, err, "pod '%s/%s' not Running", namespace, pod)
	}
}

func cleanupZoneResources(t *testing.T, argocdNamespace string, argocdOptions *terrak8s.KubectlOptions, clusterOptions *terrak8s.KubectlOptions) {
	aAppsOpts := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, AAppsNamespace)
	bAppsOpts := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, BAppsNamespace)

	_, _ = terrak8s.RunKubectlAndGetOutputE(t, aAppsOpts, "delete", "application", AAppsApplicationName, "--ignore-not-found")
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, bAppsOpts, "delete", "application", BAppsApplicationName, "--ignore-not-found")
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocdOptions, "delete", "application", ZoneApplicationName, "-n", argocdNamespace, "--ignore-not-found")
}
