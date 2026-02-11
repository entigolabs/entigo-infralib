package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
)

const (
	//postgresql
	PostgresqlConfigurationName = "platform-apis-postgresql"
	PostgresqlInstanceName      = "postgresql-instance-test"
	DatabaseFunctionName        = "platform-apis-database-fn"
	FunctionKind                = "function.pkg.crossplane.io"
	PostgresqlInstanceKind      = "postgresqlinstance.database.entigo.com"
	RdsInstanceKind             = "instance.rds.aws.m.upbound.io"
	PostgresqlDatabaseName      = "database-test"
	PostgresqlDatabaseKind      = "postgresqldatabase.database.entigo.com"
	SqlDatabaseKind             = "database.postgresql.sql.m.crossplane.io"
	SqlGrantKind                = "grant.postgresql.sql.m.crossplane.io"
	PostgresqlUserName          = "user-test"
	PostgresqlUserKind          = "postgresqluser.database.entigo.com"
	SqlRoleKind                 = "role.postgresql.sql.m.crossplane.io"
	PostgresqlUserSpecName      = "user_test"

	//zone
	AppProjectName        = "zone"
	NamespaceZoneLabelKey = "app.kubernetes.io/name"
	ZoneApplicationName   = "app-of-zone"
	ZoneAName             = "a"
	ZoneBName             = "b"
	ZoneConfigurationName = "platform-apis-zone"
	ZoneKind              = "zone.tenancy.entigo.com"

	//common
	ConfigurationKind = "configuration.pkg.crossplane.io"
	NamespaceName     = "test-namespace"
)

func TestK8sPlatformApisAWSBiz(t *testing.T) {
	testK8sPlatformApis(t, "aws", "biz")
}

// func TestK8sPlatformApisAWSPri(t *testing.T) {
// 	testK8sPlatformApis(t, "aws", "pri")
// }

func testK8sPlatformApis(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, _ := k8s.CheckKubectlConnection(t, cloudName, envName)

	argocdNamespace := fmt.Sprintf("argocd-%s", envName)
	argocdOptions := terrak8s.NewKubectlOptions(kubectlOptions.ContextName, kubectlOptions.ConfigPath, argocdNamespace)

	clusterOptions := terrak8s.NewKubectlOptions(kubectlOptions.ContextName, kubectlOptions.ConfigPath, "")

	testPlatformApisZone(t, argocdNamespace, clusterOptions, argocdOptions)
	testPlatformApisPostgresql(t, argocdNamespace, clusterOptions)
}

//---- POSTGRESQL TESTS ----

func testPlatformApisPostgresql(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions) {

	defer func() {
		if t.Failed() {
			fmt.Printf("[%s] Cleanup: skipping cleanup due to test failure\n", argocdNamespace)
			return
		}
		fmt.Printf("[%s] Cleanup: deleting test resources\n", argocdNamespace)
		nsOptions := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, NamespaceName)

		fmt.Printf("[%s] Cleanup: deleting PostgreSQL User '%s'\n", argocdNamespace, PostgresqlUserName)
		_, _ = terrak8s.RunKubectlAndGetOutputE(t, nsOptions, "delete", PostgresqlUserKind, PostgresqlUserName, "-n", NamespaceName, "--ignore-not-found")

		fmt.Printf("[%s] Cleanup: deleting PostgreSQL Database '%s'\n", argocdNamespace, PostgresqlDatabaseName)
		_, _ = terrak8s.RunKubectlAndGetOutputE(t, nsOptions, "delete", PostgresqlDatabaseKind, PostgresqlDatabaseName, "-n", NamespaceName, "--ignore-not-found")

		fmt.Printf("[%s] Cleanup: disabling deletionProtection on PostgreSQL Instance '%s'\n", argocdNamespace, PostgresqlInstanceName)
		_, _ = terrak8s.RunKubectlAndGetOutputE(t, nsOptions, "patch", PostgresqlInstanceKind, PostgresqlInstanceName, "-n", NamespaceName, "--type", "merge", "-p", `{"spec":{"deletionProtection":false}}`)

		rdsName, err := terrak8s.RunKubectlAndGetOutputE(t, nsOptions, "get", RdsInstanceKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlInstanceName), "-o", "jsonpath={.items[0].metadata.name}")
		if err == nil && rdsName != "" {
			fmt.Printf("[%s] Cleanup: disabling deletionProtection on RDS Instance '%s'\n", argocdNamespace, rdsName)
			_, _ = terrak8s.RunKubectlAndGetOutputE(t, nsOptions, "patch", RdsInstanceKind, rdsName, "-n", NamespaceName, "--type", "merge", "-p", `{"spec":{"forProvider":{"deletionProtection":false,"skipFinalSnapshot":true}}}`)
		}

		time.Sleep(30 * time.Second)

		fmt.Printf("[%s] Cleanup: deleting PostgreSQL Instance '%s'\n", argocdNamespace, PostgresqlInstanceName)
		_, _ = terrak8s.RunKubectlAndGetOutputE(t, nsOptions, "delete", PostgresqlInstanceKind, PostgresqlInstanceName, "-n", NamespaceName, "--ignore-not-found")

		fmt.Printf("[%s] Cleanup: waiting for resources to be deleted\n", argocdNamespace)
		_, _ = retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for PostgreSQL Instance deletion", argocdNamespace), 60, 10*time.Second, func() (string, error) {
			output, err := terrak8s.RunKubectlAndGetOutputE(t, nsOptions, "get", PostgresqlInstanceKind, PostgresqlInstanceName, "-n", NamespaceName, "--ignore-not-found")
			if err != nil {
				return "", err
			}
			if output != "" {
				return "", fmt.Errorf("PostgreSQL Instance '%s' still exists", PostgresqlInstanceName)
			}
			return "deleted", nil
		})

		fmt.Printf("[%s] Cleanup: deleting namespace '%s'\n", argocdNamespace, NamespaceName)
		_, _ = terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "delete", "namespace", NamespaceName, "--ignore-not-found")

		fmt.Printf("[%s] Cleanup: done\n", argocdNamespace)
	}()

	test0PlatformApisPostgresqlConfigurationDeployed(t, argocdNamespace, clusterOptions)
	test1PlatformApisDatabaseFunctionDeployed(t, argocdNamespace, clusterOptions)
	test2TestNamespaceCreated(t, argocdNamespace, clusterOptions)
	namespaceOptions := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, NamespaceName)
	test3PostgresqlInstanceApplied(t, argocdNamespace, namespaceOptions)
	test4PostgresqlInstanceSyncedAndReady(t, argocdNamespace, namespaceOptions)
	test5RdsInstanceSyncedAndReady(t, argocdNamespace, namespaceOptions)
	test6RdsInstanceFieldsVerified(t, argocdNamespace, namespaceOptions)
	test7PostgresqlDatabaseApplied(t, argocdNamespace, namespaceOptions)
	test8PostgresqlDatabaseSyncedAndReady(t, argocdNamespace, namespaceOptions)
	test9SqlDatabaseSyncedAndReady(t, argocdNamespace, namespaceOptions)
	test10SqlDatabaseFieldsVerified(t, argocdNamespace, namespaceOptions)
	test11SqlGrantSyncedAndReady(t, argocdNamespace, namespaceOptions)
	test12PostgresqlUserApplied(t, argocdNamespace, namespaceOptions)
	test13PostgresqlUserSyncedAndReady(t, argocdNamespace, namespaceOptions)
	test14SqlRoleSyncedAndReady(t, argocdNamespace, namespaceOptions)
	test15SqlRoleExternalNameVerified(t, argocdNamespace, namespaceOptions)
	test16UserSqlGrantSyncedAndReady(t, argocdNamespace, namespaceOptions)
}

func test0PlatformApisPostgresqlConfigurationDeployed(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 0: Waiting for Crossplane Configuration '%s' to be Healthy and Installed\n", argocdNamespace, PostgresqlConfigurationName)
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
	fmt.Printf("[%s] Step 0: PASSED - Configuration '%s' is Healthy and Installed\n", argocdNamespace, PostgresqlConfigurationName)
}

func test1PlatformApisDatabaseFunctionDeployed(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 1: Waiting for Crossplane Function '%s' to be Healthy and Installed\n", argocdNamespace, DatabaseFunctionName)
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
	fmt.Printf("[%s] Step 1: PASSED - Function '%s' is Healthy and Installed\n", argocdNamespace, DatabaseFunctionName)
}

func test2TestNamespaceCreated(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 2: Creating namespace '%s'\n", argocdNamespace, NamespaceName)
	_, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "create", "namespace", NamespaceName)
	if err != nil {
		_, err = terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", "namespace", NamespaceName)
		require.NoError(t, err, fmt.Sprintf("[%s] Namespace '%s' not found", argocdNamespace, NamespaceName))
	}
	fmt.Printf("[%s] Step 2: PASSED - Namespace created\n", argocdNamespace)
}

func test3PostgresqlInstanceApplied(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 3: Applying PostgreSQL Instance '%s' to namespace '%s'\n", argocdNamespace, PostgresqlInstanceName, NamespaceName)
	_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "apply", "-f", "./templates/postgresql_test_instance.yaml", "-n", NamespaceName)
	require.NoError(t, err, fmt.Sprintf("[%s] Applying PostgreSQL error", argocdNamespace))
	fmt.Printf("[%s] Step 3: PASSED - PostgreSQL applied\n", argocdNamespace)
}

func test4PostgresqlInstanceSyncedAndReady(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 4: Waiting for PostgreSQL Instance '%s' to be Synced and Ready\n", argocdNamespace, PostgresqlInstanceName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for PostgreSQL Instance '%s' to be Synced and Ready", argocdNamespace, PostgresqlInstanceName), 60, 10*time.Second, func() (string, error) {
		syncStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", PostgresqlInstanceKind, PostgresqlInstanceName, "-o", `jsonpath={.status.conditions[?(@.type=="Synced")].status}`)
		if err != nil {
			return "", err
		}
		if syncStatus != "True" {
			return "", fmt.Errorf("PostgreSQL Instance '%s' not synced yet, condition: %s", PostgresqlInstanceName, syncStatus)
		}
		readyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", PostgresqlInstanceKind, PostgresqlInstanceName, "-o", `jsonpath={.status.conditions[?(@.type=="Ready")].status}`)
		if err != nil {
			return "", err
		}
		if readyStatus != "True" {
			return "", fmt.Errorf("PostgreSQL Instance '%s' not ready yet, condition: %s", PostgresqlInstanceName, readyStatus)
		}
		return "Synced+Ready", nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] PostgreSQL Instance '%s' failed to become Synced and Ready", argocdNamespace, PostgresqlInstanceName))
	fmt.Printf("[%s] Step 4: PASSED - PostgreSQL Instance '%s' is Synced and Ready\n", argocdNamespace, PostgresqlInstanceName)
}

func test5RdsInstanceSyncedAndReady(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 5: Waiting for RDS Instance related to '%s' to be Synced and Ready\n", argocdNamespace, PostgresqlInstanceName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for RDS Instance related to '%s'", argocdNamespace, PostgresqlInstanceName), 60, 10*time.Second, func() (string, error) {
		rdsName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", RdsInstanceKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlInstanceName), "-o", "jsonpath={.items[0].metadata.name}")
		if err != nil {
			return "", err
		}
		if rdsName == "" {
			return "", fmt.Errorf("no RDS Instance found for composite '%s'", PostgresqlInstanceName)
		}
		syncStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", RdsInstanceKind, rdsName, "-o", `jsonpath={.status.conditions[?(@.type=="Synced")].status}`)
		if err != nil {
			return "", err
		}
		if syncStatus != "True" {
			return "", fmt.Errorf("RDS Instance '%s' not synced yet, condition: %s", rdsName, syncStatus)
		}
		readyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", RdsInstanceKind, rdsName, "-o", `jsonpath={.status.conditions[?(@.type=="Ready")].status}`)
		if err != nil {
			return "", err
		}
		if readyStatus != "True" {
			return "", fmt.Errorf("RDS Instance '%s' not ready yet, condition: %s", rdsName, readyStatus)
		}
		return "Synced+Ready", nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] RDS Instance for '%s' failed to become Synced and Ready", argocdNamespace, PostgresqlInstanceName))
	fmt.Printf("[%s] Step 5: PASSED - RDS Instance is Synced and Ready\n", argocdNamespace)
}

func test6RdsInstanceFieldsVerified(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 6: Verifying RDS Instance fields for '%s'\n", argocdNamespace, PostgresqlInstanceName)
	rdsName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", RdsInstanceKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlInstanceName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to find RDS Instance", argocdNamespace))
	require.NotEmpty(t, rdsName, fmt.Sprintf("[%s] No RDS Instance found for composite '%s'", argocdNamespace, PostgresqlInstanceName))

	allocatedStorage, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", RdsInstanceKind, rdsName, "-o", "jsonpath={.spec.forProvider.allocatedStorage}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get allocatedStorage", argocdNamespace))
	require.Equal(t, "20", allocatedStorage, fmt.Sprintf("[%s] RDS Instance '%s' allocatedStorage mismatch", argocdNamespace, rdsName))

	engineVersion, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", RdsInstanceKind, rdsName, "-o", "jsonpath={.spec.forProvider.engineVersion}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get engineVersion", argocdNamespace))
	require.Equal(t, "17.2", engineVersion, fmt.Sprintf("[%s] RDS Instance '%s' engineVersion mismatch", argocdNamespace, rdsName))

	instanceClass, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", RdsInstanceKind, rdsName, "-o", "jsonpath={.spec.forProvider.instanceClass}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get instanceClass", argocdNamespace))
	require.Equal(t, "db.t3.micro", instanceClass, fmt.Sprintf("[%s] RDS Instance '%s' instanceClass mismatch", argocdNamespace, rdsName))

	fmt.Printf("[%s] Step 6: PASSED - RDS Instance fields verified (allocatedStorage=20, engineVersion=17.2, instanceClass=db.t3.micro)\n", argocdNamespace)
}

func test7PostgresqlDatabaseApplied(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 7: Applying PostgreSQL Database '%s' to namespace '%s'\n", argocdNamespace, PostgresqlDatabaseName, NamespaceName)
	_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "apply", "-f", "./templates/postgresql_test_database.yaml", "-n", NamespaceName)
	require.NoError(t, err, fmt.Sprintf("[%s] Applying PostgreSQL Database error", argocdNamespace))
	fmt.Printf("[%s] Step 7: PASSED - PostgreSQL Database applied\n", argocdNamespace)
}

func test8PostgresqlDatabaseSyncedAndReady(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 8: Waiting for PostgreSQL Database '%s' to be Synced and Ready\n", argocdNamespace, PostgresqlDatabaseName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for PostgreSQL Database '%s' to be Synced and Ready", argocdNamespace, PostgresqlDatabaseName), 60, 10*time.Second, func() (string, error) {
		syncStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", PostgresqlDatabaseKind, PostgresqlDatabaseName, "-o", `jsonpath={.status.conditions[?(@.type=="Synced")].status}`)
		if err != nil {
			return "", err
		}
		if syncStatus != "True" {
			return "", fmt.Errorf("PostgreSQL Database '%s' not synced yet, condition: %s", PostgresqlDatabaseName, syncStatus)
		}
		readyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", PostgresqlDatabaseKind, PostgresqlDatabaseName, "-o", `jsonpath={.status.conditions[?(@.type=="Ready")].status}`)
		if err != nil {
			return "", err
		}
		if readyStatus != "True" {
			return "", fmt.Errorf("PostgreSQL Database '%s' not ready yet, condition: %s", PostgresqlDatabaseName, readyStatus)
		}
		return "Synced+Ready", nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] PostgreSQL Database '%s' failed to become Synced and Ready", argocdNamespace, PostgresqlDatabaseName))
	fmt.Printf("[%s] Step 8: PASSED - PostgreSQL Database '%s' is Synced and Ready\n", argocdNamespace, PostgresqlDatabaseName)
}

func test9SqlDatabaseSyncedAndReady(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 9: Waiting for SQL Database resource related to '%s' to be Synced and Ready\n", argocdNamespace, PostgresqlDatabaseName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for SQL Database related to '%s'", argocdNamespace, PostgresqlDatabaseName), 60, 10*time.Second, func() (string, error) {
		dbName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlDatabaseName), "-o", "jsonpath={.items[0].metadata.name}")
		if err != nil {
			return "", err
		}
		if dbName == "" {
			return "", fmt.Errorf("no SQL Database found for composite '%s'", PostgresqlDatabaseName)
		}
		syncStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", `jsonpath={.status.conditions[?(@.type=="Synced")].status}`)
		if err != nil {
			return "", err
		}
		if syncStatus != "True" {
			return "", fmt.Errorf("SQL Database '%s' not synced yet, condition: %s", dbName, syncStatus)
		}
		readyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", `jsonpath={.status.conditions[?(@.type=="Ready")].status}`)
		if err != nil {
			return "", err
		}
		if readyStatus != "True" {
			return "", fmt.Errorf("SQL Database '%s' not ready yet, condition: %s", dbName, readyStatus)
		}
		return "Synced+Ready", nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] SQL Database for '%s' failed to become Synced and Ready", argocdNamespace, PostgresqlDatabaseName))
	fmt.Printf("[%s] Step 9: PASSED - SQL Database is Synced and Ready\n", argocdNamespace)
}

func test10SqlDatabaseFieldsVerified(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 10: Verifying SQL Database fields for '%s'\n", argocdNamespace, PostgresqlDatabaseName)
	dbName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlDatabaseName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to find SQL Database", argocdNamespace))
	require.NotEmpty(t, dbName, fmt.Sprintf("[%s] No SQL Database found for composite '%s'", argocdNamespace, PostgresqlDatabaseName))

	encoding, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.encoding}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get encoding", argocdNamespace))
	require.Equal(t, "UTF8", encoding, fmt.Sprintf("[%s] SQL Database '%s' encoding mismatch", argocdNamespace, dbName))

	lcCType, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.lcCType}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get lcCType", argocdNamespace))
	require.Equal(t, "et_EE.UTF-8", lcCType, fmt.Sprintf("[%s] SQL Database '%s' lcCType mismatch", argocdNamespace, dbName))

	lcCollate, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.lcCollate}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get lcCollate", argocdNamespace))
	require.Equal(t, "et_EE.UTF-8", lcCollate, fmt.Sprintf("[%s] SQL Database '%s' lcCollate mismatch", argocdNamespace, dbName))

	template, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.template}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get template", argocdNamespace))
	require.Equal(t, "template0", template, fmt.Sprintf("[%s] SQL Database '%s' template mismatch", argocdNamespace, dbName))

	fmt.Printf("[%s] Step 10: PASSED - SQL Database fields verified (encoding=UTF8, lcCType=et_EE.UTF-8, lcCollate=et_EE.UTF-8, template=template0)\n", argocdNamespace)
}

func test11SqlGrantSyncedAndReady(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 11: Waiting for SQL Grant resource related to '%s' to be Synced and Ready\n", argocdNamespace, PostgresqlDatabaseName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for SQL Grant related to '%s'", argocdNamespace, PostgresqlDatabaseName), 60, 10*time.Second, func() (string, error) {
		grantName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlDatabaseName), "-o", "jsonpath={.items[0].metadata.name}")
		if err != nil {
			return "", err
		}
		if grantName == "" {
			return "", fmt.Errorf("no SQL Grant found for composite '%s'", PostgresqlDatabaseName)
		}
		syncStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, grantName, "-o", `jsonpath={.status.conditions[?(@.type=="Synced")].status}`)
		if err != nil {
			return "", err
		}
		if syncStatus != "True" {
			return "", fmt.Errorf("SQL Grant '%s' not synced yet, condition: %s", grantName, syncStatus)
		}
		readyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, grantName, "-o", `jsonpath={.status.conditions[?(@.type=="Ready")].status}`)
		if err != nil {
			return "", err
		}
		if readyStatus != "True" {
			return "", fmt.Errorf("SQL Grant '%s' not ready yet, condition: %s", grantName, readyStatus)
		}
		return "Synced+Ready", nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] SQL Grant for '%s' failed to become Synced and Ready", argocdNamespace, PostgresqlDatabaseName))
	fmt.Printf("[%s] Step 11: PASSED - SQL Grant is Synced and Ready\n", argocdNamespace)
}

func test12PostgresqlUserApplied(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 12: Applying PostgreSQL User '%s' to namespace '%s'\n", argocdNamespace, PostgresqlUserName, NamespaceName)
	_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "apply", "-f", "./templates/postgresql_test_user.yaml", "-n", NamespaceName)
	require.NoError(t, err, fmt.Sprintf("[%s] Applying PostgreSQL User error", argocdNamespace))
	fmt.Printf("[%s] Step 12: PASSED - PostgreSQL User applied\n", argocdNamespace)
}

func test13PostgresqlUserSyncedAndReady(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 13: Waiting for PostgreSQL User '%s' to be Synced and Ready\n", argocdNamespace, PostgresqlUserName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for PostgreSQL User '%s' to be Synced and Ready", argocdNamespace, PostgresqlUserName), 60, 10*time.Second, func() (string, error) {
		syncStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", PostgresqlUserKind, PostgresqlUserName, "-o", `jsonpath={.status.conditions[?(@.type=="Synced")].status}`)
		if err != nil {
			return "", err
		}
		if syncStatus != "True" {
			return "", fmt.Errorf("PostgreSQL User '%s' not synced yet, condition: %s", PostgresqlUserName, syncStatus)
		}
		readyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", PostgresqlUserKind, PostgresqlUserName, "-o", `jsonpath={.status.conditions[?(@.type=="Ready")].status}`)
		if err != nil {
			return "", err
		}
		if readyStatus != "True" {
			return "", fmt.Errorf("PostgreSQL User '%s' not ready yet, condition: %s", PostgresqlUserName, readyStatus)
		}
		return "Synced+Ready", nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] PostgreSQL User '%s' failed to become Synced and Ready", argocdNamespace, PostgresqlUserName))
	fmt.Printf("[%s] Step 13: PASSED - PostgreSQL User '%s' is Synced and Ready\n", argocdNamespace, PostgresqlUserName)
}

func test14SqlRoleSyncedAndReady(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 14: Waiting for SQL Role related to '%s' to be Synced and Ready\n", argocdNamespace, PostgresqlUserName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for SQL Role related to '%s'", argocdNamespace, PostgresqlUserName), 60, 10*time.Second, func() (string, error) {
		roleName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlUserName), "-o", "jsonpath={.items[0].metadata.name}")
		if err != nil {
			return "", err
		}
		if roleName == "" {
			return "", fmt.Errorf("no SQL Role found for composite '%s'", PostgresqlUserName)
		}
		syncStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", `jsonpath={.status.conditions[?(@.type=="Synced")].status}`)
		if err != nil {
			return "", err
		}
		if syncStatus != "True" {
			return "", fmt.Errorf("SQL Role '%s' not synced yet, condition: %s", roleName, syncStatus)
		}
		readyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", `jsonpath={.status.conditions[?(@.type=="Ready")].status}`)
		if err != nil {
			return "", err
		}
		if readyStatus != "True" {
			return "", fmt.Errorf("SQL Role '%s' not ready yet, condition: %s", roleName, readyStatus)
		}
		return "Synced+Ready", nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] SQL Role for '%s' failed to become Synced and Ready", argocdNamespace, PostgresqlUserName))
	fmt.Printf("[%s] Step 14: PASSED - SQL Role is Synced and Ready\n", argocdNamespace)
}

func test15SqlRoleExternalNameVerified(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 15: Verifying SQL Role crossplane.io/external-name annotation for '%s'\n", argocdNamespace, PostgresqlUserName)
	roleName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlUserName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to find SQL Role", argocdNamespace))
	require.NotEmpty(t, roleName, fmt.Sprintf("[%s] No SQL Role found for composite '%s'", argocdNamespace, PostgresqlUserName))

	externalName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", `jsonpath={.metadata.annotations.crossplane\.io/external-name}`)
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get crossplane.io/external-name annotation", argocdNamespace))
	require.Equal(t, PostgresqlUserSpecName, externalName, fmt.Sprintf("[%s] SQL Role '%s' crossplane.io/external-name expected '%s', got '%s'", argocdNamespace, roleName, PostgresqlUserSpecName, externalName))

	fmt.Printf("[%s] Step 15: PASSED - SQL Role crossplane.io/external-name=%s\n", argocdNamespace, PostgresqlUserSpecName)
}

func test16UserSqlGrantSyncedAndReady(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 16: Waiting for SQL Grant related to user '%s' to be Synced and Ready\n", argocdNamespace, PostgresqlUserName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for SQL Grant related to user '%s'", argocdNamespace, PostgresqlUserName), 60, 10*time.Second, func() (string, error) {
		grantName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlUserName), "-o", "jsonpath={.items[0].metadata.name}")
		if err != nil {
			return "", err
		}
		if grantName == "" {
			return "", fmt.Errorf("no SQL Grant found for composite '%s'", PostgresqlUserName)
		}
		syncStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, grantName, "-o", `jsonpath={.status.conditions[?(@.type=="Synced")].status}`)
		if err != nil {
			return "", err
		}
		if syncStatus != "True" {
			return "", fmt.Errorf("SQL Grant '%s' not synced yet, condition: %s", grantName, syncStatus)
		}
		readyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, grantName, "-o", `jsonpath={.status.conditions[?(@.type=="Ready")].status}`)
		if err != nil {
			return "", err
		}
		if readyStatus != "True" {
			return "", fmt.Errorf("SQL Grant '%s' not ready yet, condition: %s", grantName, readyStatus)
		}
		return "Synced+Ready", nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] SQL Grant for user '%s' failed to become Synced and Ready", argocdNamespace, PostgresqlUserName))
	fmt.Printf("[%s] Step 16: PASSED - User SQL Grant is Synced and Ready\n", argocdNamespace)
}

//---- ZONE TESTS ----

func testPlatformApisZone(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions, argocdOptions *terrak8s.KubectlOptions) {

	defer func() {
		if t.Failed() {
			fmt.Printf("[%s] Cleanup: skipping cleanup due to test failure\n", argocdNamespace)
			return
		}
		fmt.Printf("[%s] Cleanup: deleting test resources\n", argocdNamespace)
		_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocdOptions, "delete", "application", ZoneApplicationName, "-n", argocdNamespace)
		_, _ = terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "delete", "namespace", NamespaceName)
		fmt.Printf("[%s] Cleanup: done\n", argocdNamespace)
	}()

	test0PlatformApisZoneConfigurationDeployed(t, argocdNamespace, clusterOptions)
	test1AppProjectExists(t, argocdNamespace, argocdOptions)
	test2ZoneApplicationApplied(t, argocdNamespace, argocdOptions)
	test3VerifyZoneApplicationName(t, argocdNamespace, argocdOptions)
	test4ZoneApplicationSynced(t, argocdNamespace, argocdOptions)
	testZoneResources(t, argocdNamespace, clusterOptions)
	test8NamespaceCreated(t, argocdNamespace, clusterOptions)
	test9NamespaceHasValidZoneLabel(t, argocdNamespace, clusterOptions)
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
	fmt.Printf("[%s] Step 9: Creating namespace '%s'\n", argocdNamespace, NamespaceName)
	_, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "create", "namespace", NamespaceName)
	if err != nil {
		_, err = terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", "namespace", NamespaceName)
		require.NoError(t, err, fmt.Sprintf("[%s] Namespace '%s' not found", argocdNamespace, NamespaceName))
	}
	fmt.Printf("[%s] Step 9: PASSED - Namespace created\n", argocdNamespace)
}

func test9NamespaceHasValidZoneLabel(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 9: Verifying namespace '%s' has label %s=%s\n", argocdNamespace, NamespaceName, NamespaceZoneLabelKey, ZoneAName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for namespace label", argocdNamespace), 30, 10*time.Second, func() (string, error) {
		label, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", "namespace", NamespaceName, "-o", "jsonpath={.metadata.labels.tenancy\\.entigo\\.com/zone}")
		if err != nil {
			return "", err
		}
		if label != ZoneAName {
			return "", fmt.Errorf("namespace label %s expected '%s', got '%s'", NamespaceZoneLabelKey, ZoneAName, label)
		}
		return label, nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] Namespace '%s' label %s != '%s'", argocdNamespace, NamespaceName, NamespaceZoneLabelKey, ZoneAName))
	fmt.Printf("[%s] Step 9: PASSED - Namespace label verified (%s=%s)\n", argocdNamespace, NamespaceZoneLabelKey, ZoneAName)
}

//apply apps in zones a and b
//check apps
//check pods
//check pods nodeselectors
//create users
//check user rights
//now cleanup all created resources... PostgresqlInstance and RdsInstance have deletionProtection... change it to false first...
