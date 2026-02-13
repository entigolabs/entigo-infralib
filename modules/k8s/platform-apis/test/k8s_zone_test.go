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
	NamespaceZoneLabelKey = "app.kubernetes.io/name"
	ZoneApplicationName   = "app-of-zones"
	ZoneAName             = "a"
	ZoneBName             = "b"
	ZoneConfigurationName = "platform-apis-zone"
	ZoneKind              = "zone.tenancy.entigo.com"
	ZoneNamespaceName     = "test-zone"
)

//---- ZONE TESTS ----

func testPlatformApisZone(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions, argocdOptions *terrak8s.KubectlOptions) {
	test0PlatformApisZoneConfigurationDeployed(t, argocdNamespace, clusterOptions)
	test1AppProjectExists(t, argocdNamespace, argocdOptions)
	test2ZoneApplicationApplied(t, argocdNamespace, argocdOptions)
	test3VerifyZoneApplicationName(t, argocdNamespace, argocdOptions)
	test4ZoneApplicationSynced(t, argocdNamespace, argocdOptions)
	testZoneResources(t, argocdNamespace, clusterOptions)
	//test8NamespaceCreated(t, argocdNamespace, clusterOptions)
	//test9NamespaceHasValidZoneLabel(t, argocdNamespace, clusterOptions)
}

func test0PlatformApisZoneConfigurationDeployed(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 0: Waiting for Crossplane Configuration '%s' to be Healthy and Installed\n", argocdNamespace, ZoneConfigurationName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for Configuration '%s'", argocdNamespace, ZoneConfigurationName), 40, 6*time.Second, func() (string, error) {
		healthyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", ConfigurationKind, ZoneConfigurationName, "-o", `jsonpath={.status.conditions[?(@.type=="Healthy")].status}`)
		if err != nil {
			return "", err
		}
		if healthyStatus != "True" {
			return "", fmt.Errorf("configuration not healthy yet, status: %s", healthyStatus)
		}
		installedStatus, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", ConfigurationKind, ZoneConfigurationName, "-o", `jsonpath={.status.conditions[?(@.type=="Installed")].status}`)
		if err != nil {
			return "", err
		}
		if installedStatus != "True" {
			return "", fmt.Errorf("configuration not installed yet, status: %s", installedStatus)
		}
		return "Healthy+Installed", nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] Crossplane Configuration '%s' not ready", argocdNamespace, ZoneConfigurationName))
	fmt.Printf("[%s] Step 0: PASSED - Configuration '%s' is Healthy and Installed\n", argocdNamespace, ZoneConfigurationName)
}

func test1AppProjectExists(t *testing.T, argocdNamespace string, argocdOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 1: Checking AppProject '%s' exists\n", argocdNamespace, AppProjectName)
	_, err := terrak8s.RunKubectlAndGetOutputE(t, argocdOptions, "get", "appproject", AppProjectName, "-n", argocdNamespace)
	require.NoError(t, err, fmt.Sprintf("[%s] AppProject '%s' not found", argocdNamespace, AppProjectName))
	fmt.Printf("[%s] Step 1: PASSED - AppProject '%s' exists\n", argocdNamespace, AppProjectName)
}

func test2ZoneApplicationApplied(t *testing.T, argocdNamespace string, argocdOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 2: Applying Application '%s'\n", argocdNamespace, ZoneApplicationName)
	_, err := terrak8s.RunKubectlAndGetOutputE(t, argocdOptions, "apply", "-f", "./templates/zone_test_application.yaml", "-n", argocdNamespace)
	require.NoError(t, err, fmt.Sprintf("[%s] Applying Application error", argocdNamespace))
	fmt.Printf("[%s] Step 2: PASSED - Application applied\n", argocdNamespace)
}

func test3VerifyZoneApplicationName(t *testing.T, argocdNamespace string, argocdOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 3: Verifying Application '%s'\n", argocdNamespace, ZoneApplicationName)
	appName, err := terrak8s.RunKubectlAndGetOutputE(t, argocdOptions, "get", "application", ZoneApplicationName, "-n", argocdNamespace, "-o", "jsonpath={.metadata.name}")
	require.NoError(t, err, fmt.Sprintf("[%s] Application not found", argocdNamespace))
	require.Equal(t, ZoneApplicationName, appName, fmt.Sprintf("[%s] Application name mismatch", argocdNamespace))
	fmt.Printf("[%s] Step 3: PASSED - Application verified (name)\n", argocdNamespace)
}

func test4ZoneApplicationSynced(t *testing.T, argocdNamespace string, argocdOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 4: Triggering sync for Application '%s'\n", argocdNamespace, ZoneApplicationName)
	_, err := terrak8s.RunKubectlAndGetOutputE(t, argocdOptions, "patch", "application", ZoneApplicationName, "-n", argocdNamespace, "--type", "merge", "-p", `{"operation":{"initiatedBy":{"username":"test"},"sync":{"revision":"HEAD"}}}`)
	require.NoError(t, err, fmt.Sprintf("[%s] Force sync Application error", argocdNamespace))

	_, err = retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for Application to sync", argocdNamespace), 30, 10*time.Second, func() (string, error) {
		output, err := terrak8s.RunKubectlAndGetOutputE(t, argocdOptions, "get", "application", ZoneApplicationName, "-n", argocdNamespace, "-o", "jsonpath={.status.sync.status}")
		if err != nil {
			return "", err
		}
		if output != "Synced" {
			return "", fmt.Errorf("application not synced yet, status: %s", output)
		}
		return output, nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] Application '%s' failed to sync", argocdNamespace, ZoneApplicationName))
	fmt.Printf("[%s] Step 4: PASSED - Application synced\n", argocdNamespace)
}

func testZoneResources(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions) {
	for _, zone := range []string{ZoneAName, ZoneBName} {

		test5ZoneResourceExists(t, argocdNamespace, clusterOptions, zone)
		test6ZoneResourceSyncedAndReady(t, argocdNamespace, clusterOptions, zone)
		test7ZoneHasNodegroupAndItIsReady(t, argocdNamespace, clusterOptions, zone)
	}
}

func test5ZoneResourceExists(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions, zone string) {
	fmt.Printf("[%s] Step 5-%s: Checking Zone '%s' exists\n", argocdNamespace, zone, zone)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for zone '%s' to appear", argocdNamespace, zone), 30, 10*time.Second, func() (string, error) {
		name, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", ZoneKind, zone, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		return name, nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] Zone '%s' not found", argocdNamespace, zone))
	fmt.Printf("[%s] Step 5-%s: PASSED - Zone '%s' exists\n", argocdNamespace, zone, zone)
}

func test6ZoneResourceSyncedAndReady(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions, zone string) {
	fmt.Printf("[%s] Step 6-%s: Waiting for Zone '%s' to be Synced and Ready\n", argocdNamespace, zone, zone)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for zone '%s' to be Synced and Ready", argocdNamespace, zone), 30, 10*time.Second, func() (string, error) {
		syncStatus, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", ZoneKind, zone, "-o", `jsonpath={.status.conditions[?(@.type=="Synced")].status}`)
		if err != nil {
			return "", err
		}
		if syncStatus != "True" {
			return "", fmt.Errorf("zone '%s' not synced yet, condition: %s", zone, syncStatus)
		}
		readyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", ZoneKind, zone, "-o", `jsonpath={.status.conditions[?(@.type=="Ready")].status}`)
		if err != nil {
			return "", err
		}
		if readyStatus != "True" {
			return "", fmt.Errorf("zone '%s' not ready yet, condition: %s", zone, readyStatus)
		}
		return "Synced+Ready", nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] Zone '%s' failed to become Synced and Ready", argocdNamespace, zone))
	fmt.Printf("[%s] Step 6-%s: PASSED - Zone '%s' is Synced and Ready\n", argocdNamespace, zone, zone)
}

func test7ZoneHasNodegroupAndItIsReady(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions, zone string) {
	fmt.Printf("[%s] Step 7-%s: Checking Zone '%s' has working NodeGroup\n", argocdNamespace, zone, zone)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for zone '%s' NodeGroup to be ready", argocdNamespace, zone), 30, 10*time.Second, func() (string, error) {
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
	require.NoError(t, err, fmt.Sprintf("[%s] Zone '%s' NodeGroup not ready", argocdNamespace, zone))
	fmt.Printf("[%s] Step 7-%s: PASSED - Zone '%s' NodeGroup is Ready\n", argocdNamespace, zone, zone)
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
