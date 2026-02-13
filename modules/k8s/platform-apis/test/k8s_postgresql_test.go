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
	PostgresqlConfigurationName = "platform-apis-postgresql"
	DatabaseFunctionName        = "platform-apis-database-fn"
	FunctionKind                = "function.pkg.crossplane.io"
	PostgresqlNamespaceName     = "test-postgresql"
)

//---- POSTGRESQL TESTS ----

func testPlatformApisPostgresql(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions) {

	defer func() {
		if t.Failed() {
			fmt.Printf("[%s] Cleanup: skipping cleanup due to test failure\n", argocdNamespace)
			return
		}
		fmt.Printf("[%s] Cleanup: deleting test resources\n", argocdNamespace)
		cleanupPostgresqlResources(t, argocdNamespace, clusterOptions)
		fmt.Printf("[%s] Cleanup: done\n", argocdNamespace)
	}()

	testPlatformApisPostgresqlConfigurationDeployed(t, argocdNamespace, clusterOptions)
	testPlatformApisDatabaseFunctionDeployed(t, argocdNamespace, clusterOptions)
	testTestNamespaceCreated(t, argocdNamespace, clusterOptions)

	namespaceOptions := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, PostgresqlNamespaceName)

	runPostgresqlInstanceTests(t, argocdNamespace, namespaceOptions)
	runPostgresqlUserTests(t, argocdNamespace, namespaceOptions)
	runPostgresqlDatabaseTests(t, argocdNamespace, namespaceOptions)
}

func testPlatformApisPostgresqlConfigurationDeployed(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Waiting for Crossplane Configuration '%s' to be Healthy and Installed\n", argocdNamespace, PostgresqlConfigurationName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for Configuration '%s'", argocdNamespace, PostgresqlConfigurationName), 40, 6*time.Second, func() (string, error) {
		healthyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", ConfigurationKind, PostgresqlConfigurationName, "-o", `jsonpath={.status.conditions[?(@.type=="Healthy")].status}`)
		if err != nil {
			return "", err
		}
		if healthyStatus != "True" {
			return "", fmt.Errorf("configuration not healthy yet, status: %s", healthyStatus)
		}
		installedStatus, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", ConfigurationKind, PostgresqlConfigurationName, "-o", `jsonpath={.status.conditions[?(@.type=="Installed")].status}`)
		if err != nil {
			return "", err
		}
		if installedStatus != "True" {
			return "", fmt.Errorf("configuration not installed yet, status: %s", installedStatus)
		}
		return "Healthy+Installed", nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] Crossplane Configuration '%s' not ready", argocdNamespace, PostgresqlConfigurationName))
	fmt.Printf("[%s] TEST PASSED - Configuration '%s' is Healthy and Installed\n", argocdNamespace, PostgresqlConfigurationName)
}

func testPlatformApisDatabaseFunctionDeployed(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Waiting for Crossplane Function '%s' to be Healthy and Installed\n", argocdNamespace, DatabaseFunctionName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for Function '%s'", argocdNamespace, DatabaseFunctionName), 40, 6*time.Second, func() (string, error) {
		healthyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", FunctionKind, DatabaseFunctionName, "-o", `jsonpath={.status.conditions[?(@.type=="Healthy")].status}`)
		if err != nil {
			return "", err
		}
		if healthyStatus != "True" {
			return "", fmt.Errorf("function not healthy yet, status: %s", healthyStatus)
		}
		installedStatus, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", FunctionKind, DatabaseFunctionName, "-o", `jsonpath={.status.conditions[?(@.type=="Installed")].status}`)
		if err != nil {
			return "", err
		}
		if installedStatus != "True" {
			return "", fmt.Errorf("function not installed yet, status: %s", installedStatus)
		}
		return "Healthy+Installed", nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] Crossplane Function '%s' not ready", argocdNamespace, DatabaseFunctionName))
	fmt.Printf("[%s] TEST PASSED - Function '%s' is Healthy and Installed\n", argocdNamespace, DatabaseFunctionName)
}

func testTestNamespaceCreated(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Creating namespace '%s'\n", argocdNamespace, PostgresqlNamespaceName)
	_, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "create", "namespace", PostgresqlNamespaceName)
	if err != nil {
		_, err = terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", "namespace", PostgresqlNamespaceName)
		require.NoError(t, err, fmt.Sprintf("[%s] Namespace '%s' not found", argocdNamespace, PostgresqlNamespaceName))
	}
	fmt.Printf("[%s] TEST PASSED - Namespace created\n", argocdNamespace)
}
