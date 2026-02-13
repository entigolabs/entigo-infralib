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
	PostgresqlInstanceName      = "postgresql-instance-test"
	DatabaseFunctionName        = "platform-apis-database-fn"
	FunctionKind                = "function.pkg.crossplane.io"
	PostgresqlInstanceKind      = "postgresqlinstance.database.entigo.com"
	RdsInstanceKind             = "instance.rds.aws.m.upbound.io"
	PostgresqlDatabaseName      = "database-test"
	PostgresqlDatabaseKind      = "postgresqldatabase.database.entigo.com"
	SqlDatabaseKind             = "database.postgresql.sql.m.crossplane.io"
	SqlGrantKind                = "grant.postgresql.sql.m.crossplane.io"
	PostgresqlAdminUserName     = "test-admin"
	PostgresqlAdminUserKind     = "postgresqluser.database.entigo.com"
	PostgresqlAdminUserSpecName = "test_admin"
	PostgresqlUserName          = "test-user"
	PostgresqlUserKind          = "postgresqluser.database.entigo.com"
	SqlRoleKind                 = "role.postgresql.sql.m.crossplane.io"
	PostgresqlNamespaceName     = "test-postgresql"
)

//---- POSTGRESQL TESTS ----

func testPlatformApisPostgresql(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions) {
	test0PlatformApisPostgresqlConfigurationDeployed(t, argocdNamespace, clusterOptions)
	test1PlatformApisDatabaseFunctionDeployed(t, argocdNamespace, clusterOptions)
	test2TestNamespaceCreated(t, argocdNamespace, clusterOptions)

	namespaceOptions := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, PostgresqlNamespaceName)

	test3PostgresqlInstanceApplied(t, argocdNamespace, namespaceOptions)
	test4PostgresqlInstanceSyncedAndReady(t, argocdNamespace, namespaceOptions)
	test5RdsInstanceSyncedAndReady(t, argocdNamespace, namespaceOptions)
	test6RdsInstanceFieldsVerified(t, argocdNamespace, namespaceOptions)
	test7AdminUserApplied(t, argocdNamespace, namespaceOptions)
	test8AdminUserSyncedAndReady(t, argocdNamespace, namespaceOptions)
	test9AdminRoleSyncedAndReady(t, argocdNamespace, namespaceOptions)
	test10AdminRoleExternalNameVerified(t, argocdNamespace, namespaceOptions)
	test13DatabaseApplied(t, argocdNamespace, namespaceOptions)
	test14DatabaseSyncedAndReady(t, argocdNamespace, namespaceOptions)
	test15DatabaseFieldsVerified(t, argocdNamespace, namespaceOptions)
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
	fmt.Printf("[%s] Step 2: Creating namespace '%s'\n", argocdNamespace, PostgresqlNamespaceName)
	_, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "create", "namespace", PostgresqlNamespaceName)
	if err != nil {
		_, err = terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", "namespace", PostgresqlNamespaceName)
		require.NoError(t, err, fmt.Sprintf("[%s] Namespace '%s' not found", argocdNamespace, PostgresqlNamespaceName))
	}
	fmt.Printf("[%s] Step 2: PASSED - Namespace created\n", argocdNamespace)
}

func test3PostgresqlInstanceApplied(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 3: Applying PostgreSQL Instance '%s' to namespace '%s'\n", argocdNamespace, PostgresqlInstanceName, PostgresqlNamespaceName)
	_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "apply", "-f", "./templates/postgresql_test_instance.yaml", "-n", PostgresqlNamespaceName)
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

func test7AdminUserApplied(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 7: Applying PostgreSQL Admin User '%s' to namespace '%s'\n", argocdNamespace, PostgresqlAdminUserName, PostgresqlNamespaceName)
	_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "apply", "-f", "./templates/postgresql_test_admin_user.yaml", "-n", PostgresqlNamespaceName)
	require.NoError(t, err, fmt.Sprintf("[%s] Applying PostgreSQL Admin User error", argocdNamespace))
	fmt.Printf("[%s] Step 7: PASSED - PostgreSQL Admin User applied\n", argocdNamespace)
}

func test8AdminUserSyncedAndReady(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 8: Waiting for PostgreSQL Admin User '%s' to be Synced and Ready\n", argocdNamespace, PostgresqlAdminUserName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for PostgreSQL Admin User '%s' to be Synced and Ready", argocdNamespace, PostgresqlAdminUserName), 60, 10*time.Second, func() (string, error) {
		syncStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", PostgresqlAdminUserKind, PostgresqlAdminUserName, "-o", `jsonpath={.status.conditions[?(@.type=="Synced")].status}`)
		if err != nil {
			return "", err
		}
		if syncStatus != "True" {
			return "", fmt.Errorf("PostgreSQL Admin User '%s' not synced yet, condition: %s", PostgresqlAdminUserName, syncStatus)
		}
		readyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", PostgresqlAdminUserKind, PostgresqlAdminUserName, "-o", `jsonpath={.status.conditions[?(@.type=="Ready")].status}`)
		if err != nil {
			return "", err
		}
		if readyStatus != "True" {
			return "", fmt.Errorf("PostgreSQL Admin User '%s' not ready yet, condition: %s", PostgresqlAdminUserName, readyStatus)
		}
		return "Synced+Ready", nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] PostgreSQL Admin User '%s' failed to become Synced and Ready", argocdNamespace, PostgresqlAdminUserName))
	fmt.Printf("[%s] Step 8: PASSED - PostgreSQL Admin User '%s' is Synced and Ready\n", argocdNamespace, PostgresqlAdminUserName)
}

func test9AdminRoleSyncedAndReady(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 9: Waiting for SQL Role related to admin user '%s' to be Synced and Ready\n", argocdNamespace, PostgresqlAdminUserName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for SQL Role related to '%s'", argocdNamespace, PostgresqlAdminUserName), 60, 10*time.Second, func() (string, error) {
		roleName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlAdminUserName), "-o", "jsonpath={.items[0].metadata.name}")
		if err != nil {
			return "", err
		}
		if roleName == "" {
			return "", fmt.Errorf("no SQL Role found for composite '%s'", PostgresqlAdminUserName)
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
	require.NoError(t, err, fmt.Sprintf("[%s] SQL Role for admin user '%s' failed to become Synced and Ready", argocdNamespace, PostgresqlAdminUserName))
	fmt.Printf("[%s] Step 9: PASSED - Admin SQL Role is Synced and Ready\n", argocdNamespace)
}

func test10AdminRoleExternalNameVerified(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 10: Verifying SQL Role crossplane.io/external-name annotation for admin user '%s'\n", argocdNamespace, PostgresqlAdminUserName)
	roleName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlAdminUserName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to find SQL Role", argocdNamespace))
	require.NotEmpty(t, roleName, fmt.Sprintf("[%s] No SQL Role found for composite '%s'", argocdNamespace, PostgresqlAdminUserName))

	externalName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", `jsonpath={.metadata.annotations.crossplane\.io/external-name}`)
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get crossplane.io/external-name annotation", argocdNamespace))
	require.Equal(t, PostgresqlAdminUserSpecName, externalName, fmt.Sprintf("[%s] SQL Role '%s' crossplane.io/external-name expected '%s', got '%s'", argocdNamespace, roleName, PostgresqlAdminUserSpecName, externalName))

	fmt.Printf("[%s] Step 10: PASSED - Admin SQL Role crossplane.io/external-name=%s\n", argocdNamespace, PostgresqlAdminUserSpecName)
}

func test13DatabaseApplied(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 13: Applying PostgreSQL Database '%s' to namespace '%s'\n", argocdNamespace, PostgresqlDatabaseName, PostgresqlNamespaceName)
	_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "apply", "-f", "./templates/postgresql_test_database.yaml", "-n", PostgresqlNamespaceName)
	require.NoError(t, err, fmt.Sprintf("[%s] Applying PostgreSQL Database error", argocdNamespace))
	fmt.Printf("[%s] Step 13: PASSED - PostgreSQL Database applied\n", argocdNamespace)
}

func test14DatabaseSyncedAndReady(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 14: Waiting for PostgreSQL Database '%s' to be Synced and Ready\n", argocdNamespace, PostgresqlDatabaseName)
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
	fmt.Printf("[%s] Step 14: PASSED - PostgreSQL Database '%s' is Synced and Ready\n", argocdNamespace, PostgresqlDatabaseName)
}

func test15DatabaseFieldsVerified(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] Step 15: Verifying SQL Database fields for '%s'\n", argocdNamespace, PostgresqlDatabaseName)
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

	fmt.Printf("[%s] Step 15: PASSED - SQL Database fields verified (encoding=UTF8, lcCType=et_EE.UTF-8, lcCollate=et_EE.UTF-8, template=template0)\n", argocdNamespace)
}

//---- CLEANUP FUNCTIONS ----

func cleanupPostgresqlResources(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions) {
	pgNsOptions := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, PostgresqlNamespaceName)

	fmt.Printf("[%s] Cleanup: deleting PostgreSQL User '%s'\n", argocdNamespace, PostgresqlUserName)
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, pgNsOptions, "delete", PostgresqlUserKind, PostgresqlUserName, "-n", PostgresqlNamespaceName, "--ignore-not-found")

	fmt.Printf("[%s] Cleanup: deleting PostgreSQL Admin User '%s'\n", argocdNamespace, PostgresqlAdminUserName)
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, pgNsOptions, "delete", PostgresqlAdminUserKind, PostgresqlAdminUserName, "-n", PostgresqlNamespaceName, "--ignore-not-found")

	fmt.Printf("[%s] Cleanup: deleting PostgreSQL Database '%s'\n", argocdNamespace, PostgresqlDatabaseName)
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, pgNsOptions, "delete", PostgresqlDatabaseKind, PostgresqlDatabaseName, "-n", PostgresqlNamespaceName, "--ignore-not-found")

	fmt.Printf("[%s] Cleanup: disabling deletionProtection on PostgreSQL Instance '%s'\n", argocdNamespace, PostgresqlInstanceName)
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, pgNsOptions, "patch", PostgresqlInstanceKind, PostgresqlInstanceName, "-n", PostgresqlNamespaceName, "--type", "merge", "-p", `{"spec":{"deletionProtection":false}}`)

	rdsName, err := terrak8s.RunKubectlAndGetOutputE(t, pgNsOptions, "get", RdsInstanceKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlInstanceName), "-o", "jsonpath={.items[0].metadata.name}")
	if err == nil && rdsName != "" {
		fmt.Printf("[%s] Cleanup: disabling deletionProtection on RDS Instance '%s'\n", argocdNamespace, rdsName)
		_, _ = terrak8s.RunKubectlAndGetOutputE(t, pgNsOptions, "patch", RdsInstanceKind, rdsName, "-n", PostgresqlNamespaceName, "--type", "merge", "-p", `{"spec":{"forProvider":{"deletionProtection":false,"skipFinalSnapshot":true}}}`)
	}

	time.Sleep(30 * time.Second)

	fmt.Printf("[%s] Cleanup: deleting PostgreSQL Instance '%s'\n", argocdNamespace, PostgresqlInstanceName)
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, pgNsOptions, "delete", PostgresqlInstanceKind, PostgresqlInstanceName, "-n", PostgresqlNamespaceName, "--ignore-not-found")

	fmt.Printf("[%s] Cleanup: waiting for resources to be deleted\n", argocdNamespace)
	_, _ = retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for PostgreSQL Instance deletion", argocdNamespace), 60, 10*time.Second, func() (string, error) {
		output, err := terrak8s.RunKubectlAndGetOutputE(t, pgNsOptions, "get", PostgresqlInstanceKind, PostgresqlInstanceName, "-n", PostgresqlNamespaceName, "--ignore-not-found")
		if err != nil {
			return "", err
		}
		if output != "" {
			return "", fmt.Errorf("PostgreSQL Instance '%s' still exists", PostgresqlInstanceName)
		}
		return "deleted", nil
	})

	fmt.Printf("[%s] Cleanup: deleting namespace '%s'\n", argocdNamespace, PostgresqlNamespaceName)
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "delete", "namespace", PostgresqlNamespaceName, "--ignore-not-found")
}
