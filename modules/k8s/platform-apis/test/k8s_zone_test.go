package test

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
)

const (
	AppProjectName        = "zone"
	NamespaceZoneLabelKey = "app.kubernetes.io/name"
	ZoneApplicationName   = "app-of-zones"
	ZoneAName             = "a"
	ZoneBName             = "b"
	ZoneConfigurationName = "platform-apis-zone"
	ZoneKind              = "zone.tenancy.entigo.com"
	ZoneNamespaceName     = "test-zone"
	TenancyFunctionName   = "platform-apis-tenancy-fn"
	FunctionKind          = "function.pkg.crossplane.io"
)

func testPlatformApisZone(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions, argocdOptions *terrak8s.KubectlOptions, signalZonesReady func(bool)) {
	setupFailed := true
	defer func() {
		if setupFailed {
			signalZonesReady(false)
		}
	}()

	test1AppProjectExists(t, argocdNamespace, argocdOptions)
	test2ZoneApplicationApplied(t, argocdNamespace, argocdOptions)
	test3VerifyZoneApplicationName(t, argocdNamespace, argocdOptions)
	test4ZoneApplicationSynced(t, argocdNamespace, argocdOptions)

	setupFailed = false
	testZoneResourcesParallel(t, clusterOptions, signalZonesReady)
	//test8NamespaceCreated(t, argocdNamespace, clusterOptions)
	//test9NamespaceHasValidZoneLabel(t, argocdNamespace, clusterOptions)
}

func test1AppProjectExists(t *testing.T, argocdNamespace string, argocdOptions *terrak8s.KubectlOptions) {
	_, err := terrak8s.RunKubectlAndGetOutputE(t, argocdOptions, "get", "appproject", AppProjectName, "-n", argocdNamespace)
	require.NoError(t, err, fmt.Sprintf("AppProject '%s' not found", AppProjectName))
}

func test2ZoneApplicationApplied(t *testing.T, argocdNamespace string, argocdOptions *terrak8s.KubectlOptions) {
	_, err := terrak8s.RunKubectlAndGetOutputE(t, argocdOptions, "apply", "-f", "./templates/zone_test_application.yaml", "-n", argocdNamespace)
	require.NoError(t, err, "applying Application error")
}

func test3VerifyZoneApplicationName(t *testing.T, argocdNamespace string, argocdOptions *terrak8s.KubectlOptions) {
	appName, err := terrak8s.RunKubectlAndGetOutputE(t, argocdOptions, "get", "application", ZoneApplicationName, "-n", argocdNamespace, "-o", "jsonpath={.metadata.name}")
	require.NoError(t, err, "Application not found")
	require.Equal(t, ZoneApplicationName, appName, "Application name mismatch")
}

func test4ZoneApplicationSynced(t *testing.T, argocdNamespace string, argocdOptions *terrak8s.KubectlOptions) {
	_, err := terrak8s.RunKubectlAndGetOutputE(t, argocdOptions, "patch", "application", ZoneApplicationName, "-n", argocdNamespace, "--type", "merge", "-p", `{"operation":{"initiatedBy":{"username":"test"},"sync":{"revision":"HEAD"}}}`)
	require.NoError(t, err, "force sync Application error")

	_, err = retry.DoWithRetryE(t, fmt.Sprintf("waiting for Application '%s' to sync", ZoneApplicationName), 30, 10*time.Second, func() (string, error) {
		output, err := terrak8s.RunKubectlAndGetOutputE(t, argocdOptions, "get", "application", ZoneApplicationName, "-n", argocdNamespace, "-o", "jsonpath={.status.sync.status}")
		if err != nil {
			return "", err
		}
		if output != "Synced" {
			return "", fmt.Errorf("application not synced yet, status: %s", output)
		}
		return output, nil
	})
	require.NoError(t, err, fmt.Sprintf("Application '%s' failed to sync", ZoneApplicationName))
}

func testZoneResourcesParallel(t *testing.T, clusterOptions *terrak8s.KubectlOptions, signalZonesReady func(bool)) {
	var readyCount atomic.Int32

	for _, zone := range []string{ZoneAName, ZoneBName} {
		zone := zone
		t.Run(fmt.Sprintf("zone-%s", zone), func(t *testing.T) {
			t.Parallel()
			defer func() {
				if t.Failed() {
					signalZonesReady(false)
				}
			}()

			test5ZoneResourceExists(t, clusterOptions, zone)
			waitSyncedAndReady(t, clusterOptions, ZoneKind, zone, 30, 10*time.Second)

			if readyCount.Add(1) == 2 {
				signalZonesReady(true)
			}

			test7ZoneHasNodegroupAndItIsReady(t, clusterOptions, zone)
		})
	}
}

func test5ZoneResourceExists(t *testing.T, clusterOptions *terrak8s.KubectlOptions, zone string) {
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for zone '%s'", zone), 30, 10*time.Second, func() (string, error) {
		name, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", ZoneKind, zone, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		return name, nil
	})
	require.NoError(t, err, fmt.Sprintf("zone '%s' not found", zone))
}

func test7ZoneHasNodegroupAndItIsReady(t *testing.T, clusterOptions *terrak8s.KubectlOptions, zone string) {
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for zone '%s' NodeGroup", zone), 30, 10*time.Second, func() (string, error) {
		count, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", "nodegroup.eks.aws.upbound.io", "-l", fmt.Sprintf("crossplane.io/composite=%s", zone), "-o", "jsonpath={.items[*].metadata.name}")
		if err != nil {
			return "", err
		}
		if count == "" {
			return "", fmt.Errorf("zone '%s' has no NodeGroups", zone)
		}
		readyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", "nodegroup.eks.aws.upbound.io", "-l", fmt.Sprintf("crossplane.io/composite=%s", zone), "-o", `jsonpath={.items[0].status.conditions[?(@.type=="Ready")].status}`)
		if err != nil {
			return "", err
		}
		if readyStatus != "True" {
			return "", fmt.Errorf("zone '%s' NodeGroup not ready yet, condition: %s", zone, readyStatus)
		}
		return readyStatus, nil
	})
	require.NoError(t, err, fmt.Sprintf("zone '%s' NodeGroup not ready", zone))
}

func test8NamespaceCreated(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 9: Creating namespace '%s'\n", argocdNamespace, ZoneNamespaceName)
	_, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "create", "namespace", ZoneNamespaceName)
	if err != nil {
		_, err = terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", "namespace", ZoneNamespaceName)
		require.NoError(t, err, fmt.Sprintf("[%s] Namespace '%s' not found", argocdNamespace, ZoneNamespaceName))
	}
	fmt.Printf("[%s] Step 9: PASSED - Namespace created\n", argocdNamespace)
}

func test9NamespaceHasValidZoneLabel(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 9: Verifying namespace '%s' has label %s=%s\n", argocdNamespace, ZoneNamespaceName, NamespaceZoneLabelKey, ZoneAName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for namespace label", argocdNamespace), 30, 10*time.Second, func() (string, error) {
		label, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", "namespace", ZoneNamespaceName, "-o", "jsonpath={.metadata.labels.tenancy\\.entigo\\.com/zone}")
		if err != nil {
			return "", err
		}
		if label != ZoneAName {
			return "", fmt.Errorf("namespace label %s expected '%s', got '%s'", NamespaceZoneLabelKey, ZoneAName, label)
		}
		return label, nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] Namespace '%s' label %s != '%s'", argocdNamespace, ZoneNamespaceName, NamespaceZoneLabelKey, ZoneAName))
	fmt.Printf("[%s] Step 9: PASSED - Namespace label verified (%s=%s)\n", argocdNamespace, NamespaceZoneLabelKey, ZoneAName)
}

//---- CLEANUP FUNCTIONS ----

func cleanupZoneResources(t *testing.T, argocdNamespace string, argocdOptions *terrak8s.KubectlOptions, clusterOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Cleanup: deleting namespace '%s'\n", argocdNamespace, ZoneNamespaceName)
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "delete", "namespace", ZoneNamespaceName, "--ignore-not-found")

	fmt.Printf("[%s] Cleanup: deleting Zone Application '%s'\n", argocdNamespace, ZoneApplicationName)
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocdOptions, "delete", "application", ZoneApplicationName, "-n", argocdNamespace, "--ignore-not-found")
}

//apply apps in zones a and b
//check apps
//check pods
//check pods nodeselectors
//create users
//check user rights
//now cleanup all created resources... PostgresqlInstance and RdsInstance have deletionProtection... change it to false first...
