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
	PostgresqlDatabaseName = "database-test"
	PostgresqlDatabaseKind = "postgresqldatabase.database.entigo.com"
	SqlDatabaseKind        = "database.postgresql.sql.m.crossplane.io"
	SqlRoleKind            = "role.postgresql.sql.m.crossplane.io"
	UsageKind              = "usage.protection.crossplane.io"
	MinimalDatabaseName    = "database-minimal-test"
	// Grant name from composition: {owner | replace "_" "-"}-to-dbadmin-grant
	DatabaseGrantExpectedName = "test-admin-to-dbadmin-grant"
	// Usage name from composition: {metadata.name}-grant-usage
	DatabaseUsageExpectedName        = PostgresqlDatabaseName + "-grant-usage"
	MinimalDatabaseUsageExpectedName = MinimalDatabaseName + "-grant-usage"
)

func runPostgresqlDatabaseTests(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {

	testDatabaseApplied(t, argocdNamespace, namespaceOptions)
	testDatabaseSyncedAndReady(t, argocdNamespace, namespaceOptions)
	testDatabaseGrantOwnerToDbadmin(t, argocdNamespace, namespaceOptions)
	testDatabaseOwnerFieldVerified(t, argocdNamespace, namespaceOptions)
	testDatabaseFieldsVerified(t, argocdNamespace, namespaceOptions)
	testDatabaseUsageVerified(t, argocdNamespace, namespaceOptions)
	testMinimalDatabaseApplied(t, argocdNamespace, namespaceOptions)
	testMinimalDatabaseSyncedAndReady(t, argocdNamespace, namespaceOptions)
	testMinimalDatabaseDefaultsVerified(t, argocdNamespace, namespaceOptions)
	testMinimalDatabaseUsageVerified(t, argocdNamespace, namespaceOptions)
	testDatabaseUsagePreventsGrantDeletion(t, argocdNamespace, namespaceOptions)
}

func testDatabaseApplied(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Applying PostgreSQL Database '%s' to namespace '%s'\n", argocdNamespace, PostgresqlDatabaseName, PostgresqlNamespaceName)
	_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "apply", "-f", "./templates/postgresql_test_database.yaml", "-n", PostgresqlNamespaceName)
	require.NoError(t, err, fmt.Sprintf("[%s] Applying PostgreSQL Database error", argocdNamespace))
	fmt.Printf("[%s] TEST PASSED - PostgreSQL Database applied\n", argocdNamespace)
}

func testDatabaseSyncedAndReady(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Waiting for PostgreSQL Database '%s' to be Synced and Ready\n", argocdNamespace, PostgresqlDatabaseName)
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
	fmt.Printf("[%s] TEST PASSED - PostgreSQL Database '%s' is Synced and Ready\n", argocdNamespace, PostgresqlDatabaseName)
}

func testDatabaseGrantOwnerToDbadmin(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Verifying Grant '%s' for owner-to-dbadmin\n", argocdNamespace, DatabaseGrantExpectedName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for Grant '%s'", argocdNamespace, DatabaseGrantExpectedName), 60, 10*time.Second, func() (string, error) {
		grantName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, DatabaseGrantExpectedName, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if grantName == "" {
			return "", fmt.Errorf("Grant '%s' not found", DatabaseGrantExpectedName)
		}
		syncStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, DatabaseGrantExpectedName, "-o", `jsonpath={.status.conditions[?(@.type=="Synced")].status}`)
		if err != nil {
			return "", err
		}
		if syncStatus != "True" {
			return "", fmt.Errorf("Grant '%s' not synced yet, condition: %s", DatabaseGrantExpectedName, syncStatus)
		}
		readyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, DatabaseGrantExpectedName, "-o", `jsonpath={.status.conditions[?(@.type=="Ready")].status}`)
		if err != nil {
			return "", err
		}
		if readyStatus != "True" {
			return "", fmt.Errorf("Grant '%s' not ready yet, condition: %s", DatabaseGrantExpectedName, readyStatus)
		}
		return "Synced+Ready", nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] Grant '%s' failed to become Synced and Ready", argocdNamespace, DatabaseGrantExpectedName))

	// Verify role = "dbadmin"
	role, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, DatabaseGrantExpectedName, "-o", "jsonpath={.spec.forProvider.role}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get Grant role", argocdNamespace))
	require.Equal(t, "dbadmin", role, fmt.Sprintf("[%s] Grant '%s' role mismatch", argocdNamespace, DatabaseGrantExpectedName))

	// Verify memberOf = "test_admin" (the owner value)
	memberOf, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, DatabaseGrantExpectedName, "-o", "jsonpath={.spec.forProvider.memberOf}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get Grant memberOf", argocdNamespace))
	require.Equal(t, PostgresqlAdminUserSpecName, memberOf, fmt.Sprintf("[%s] Grant '%s' memberOf mismatch", argocdNamespace, DatabaseGrantExpectedName))

	fmt.Printf("[%s] TEST PASSED - Grant '%s' verified (role=dbadmin, memberOf=%s)\n", argocdNamespace, DatabaseGrantExpectedName, PostgresqlAdminUserSpecName)
}

func testDatabaseOwnerFieldVerified(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Verifying SQL Database owner field for '%s'\n", argocdNamespace, PostgresqlDatabaseName)
	dbName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlDatabaseName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to find SQL Database", argocdNamespace))
	require.NotEmpty(t, dbName, fmt.Sprintf("[%s] No SQL Database found for composite '%s'", argocdNamespace, PostgresqlDatabaseName))

	owner, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.owner}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get owner", argocdNamespace))
	require.Equal(t, PostgresqlAdminUserSpecName, owner, fmt.Sprintf("[%s] SQL Database '%s' owner mismatch", argocdNamespace, dbName))

	fmt.Printf("[%s] TEST PASSED - SQL Database owner=%s\n", argocdNamespace, PostgresqlAdminUserSpecName)
}

func testDatabaseFieldsVerified(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Verifying SQL Database fields for '%s'\n", argocdNamespace, PostgresqlDatabaseName)
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

	fmt.Printf("[%s] TEST PASSED - SQL Database fields verified (encoding=UTF8, lcCType=et_EE.UTF-8, lcCollate=et_EE.UTF-8, template=template0)\n", argocdNamespace)
}

// ---- MINIMAL DATABASE TESTS (no optional fields) ----

func testMinimalDatabaseApplied(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Applying minimal PostgreSQL Database '%s' to namespace '%s'\n", argocdNamespace, MinimalDatabaseName, PostgresqlNamespaceName)
	_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "apply", "-f", "./templates/postgresql_test_database_minimal.yaml", "-n", PostgresqlNamespaceName)
	require.NoError(t, err, fmt.Sprintf("[%s] Applying minimal PostgreSQL Database error", argocdNamespace))
	fmt.Printf("[%s] TEST PASSED - Minimal PostgreSQL Database applied\n", argocdNamespace)
}

func testMinimalDatabaseSyncedAndReady(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Waiting for minimal PostgreSQL Database '%s' to be Synced and Ready\n", argocdNamespace, MinimalDatabaseName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for minimal PostgreSQL Database '%s' to be Synced and Ready", argocdNamespace, MinimalDatabaseName), 60, 10*time.Second, func() (string, error) {
		syncStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", PostgresqlDatabaseKind, MinimalDatabaseName, "-o", `jsonpath={.status.conditions[?(@.type=="Synced")].status}`)
		if err != nil {
			return "", err
		}
		if syncStatus != "True" {
			return "", fmt.Errorf("minimal PostgreSQL Database '%s' not synced yet, condition: %s", MinimalDatabaseName, syncStatus)
		}
		readyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", PostgresqlDatabaseKind, MinimalDatabaseName, "-o", `jsonpath={.status.conditions[?(@.type=="Ready")].status}`)
		if err != nil {
			return "", err
		}
		if readyStatus != "True" {
			return "", fmt.Errorf("minimal PostgreSQL Database '%s' not ready yet, condition: %s", MinimalDatabaseName, readyStatus)
		}
		return "Synced+Ready", nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] Minimal PostgreSQL Database '%s' failed to become Synced and Ready", argocdNamespace, MinimalDatabaseName))
	fmt.Printf("[%s] TEST PASSED - Minimal PostgreSQL Database '%s' is Synced and Ready\n", argocdNamespace, MinimalDatabaseName)
}

func testMinimalDatabaseDefaultsVerified(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Verifying minimal SQL Database fields for '%s'\n", argocdNamespace, MinimalDatabaseName)
	dbName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", MinimalDatabaseName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to find SQL Database for minimal", argocdNamespace))
	require.NotEmpty(t, dbName, fmt.Sprintf("[%s] No SQL Database found for composite '%s'", argocdNamespace, MinimalDatabaseName))

	// Owner should still be set (test-user for minimal database)
	owner, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.owner}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get owner", argocdNamespace))
	require.Equal(t, PostgresqlRegularUserName, owner, fmt.Sprintf("[%s] SQL Database '%s' owner mismatch", argocdNamespace, dbName))

	fmt.Printf("[%s] TEST PASSED - Minimal SQL Database owner=%s\n", argocdNamespace, PostgresqlRegularUserName)
}

func testDatabaseUsageVerified(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Verifying Usage '%s' protects Grant from premature deletion\n", argocdNamespace, DatabaseUsageExpectedName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for Usage '%s'", argocdNamespace, DatabaseUsageExpectedName), 30, 10*time.Second, func() (string, error) {
		name, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, DatabaseUsageExpectedName, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if name == "" {
			return "", fmt.Errorf("Usage '%s' not found", DatabaseUsageExpectedName)
		}
		return name, nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] Usage '%s' not found", argocdNamespace, DatabaseUsageExpectedName))

	// Verify spec.of references the Grant
	ofKind, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, DatabaseUsageExpectedName, "-o", "jsonpath={.spec.of.kind}")
	require.NoError(t, err)
	require.Equal(t, "Grant", ofKind)

	ofName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, DatabaseUsageExpectedName, "-o", "jsonpath={.spec.of.resourceRef.name}")
	require.NoError(t, err)
	require.Equal(t, DatabaseGrantExpectedName, ofName)

	// Verify spec.by references the Database
	byKind, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, DatabaseUsageExpectedName, "-o", "jsonpath={.spec.by.kind}")
	require.NoError(t, err)
	require.Equal(t, "Database", byKind)

	byName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, DatabaseUsageExpectedName, "-o", "jsonpath={.spec.by.resourceRef.name}")
	require.NoError(t, err)
	require.Equal(t, PostgresqlDatabaseName, byName)

	// Verify replayDeletion is enabled
	replayDeletion, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, DatabaseUsageExpectedName, "-o", "jsonpath={.spec.replayDeletion}")
	require.NoError(t, err)
	require.Equal(t, "true", replayDeletion)

	fmt.Printf("[%s] TEST PASSED - Usage '%s' verified (of=Grant/%s, by=Database/%s, replayDeletion=true)\n", argocdNamespace, DatabaseUsageExpectedName, DatabaseGrantExpectedName, PostgresqlDatabaseName)
}

func testMinimalDatabaseUsageVerified(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Verifying Usage '%s' for minimal database\n", argocdNamespace, MinimalDatabaseUsageExpectedName)

	// Minimal database owner is "test-user", so grant name is "test-user-to-dbadmin-grant"
	expectedGrantName := "test-user-to-dbadmin-grant"

	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for Usage '%s'", argocdNamespace, MinimalDatabaseUsageExpectedName), 30, 10*time.Second, func() (string, error) {
		name, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, MinimalDatabaseUsageExpectedName, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if name == "" {
			return "", fmt.Errorf("Usage '%s' not found", MinimalDatabaseUsageExpectedName)
		}
		return name, nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] Usage '%s' not found", argocdNamespace, MinimalDatabaseUsageExpectedName))

	ofName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, MinimalDatabaseUsageExpectedName, "-o", "jsonpath={.spec.of.resourceRef.name}")
	require.NoError(t, err)
	require.Equal(t, expectedGrantName, ofName)

	byName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, MinimalDatabaseUsageExpectedName, "-o", "jsonpath={.spec.by.resourceRef.name}")
	require.NoError(t, err)
	require.Equal(t, MinimalDatabaseName, byName)

	fmt.Printf("[%s] TEST PASSED - Minimal database Usage '%s' verified (of=Grant/%s, by=Database/%s)\n", argocdNamespace, MinimalDatabaseUsageExpectedName, expectedGrantName, MinimalDatabaseName)
}

// testDatabaseUsagePreventsGrantDeletion verifies that the Usage resource blocks
// deletion of the Grant while the Database still exists.
func testDatabaseUsagePreventsGrantDeletion(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Verifying Usage prevents Grant '%s' from being deleted while Database exists\n", argocdNamespace, DatabaseGrantExpectedName)

	// Attempt to delete the Grant directly - Usage should block this
	output, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "delete", SqlGrantKind, DatabaseGrantExpectedName, "--wait=false")
	fmt.Printf("[%s] Delete attempt output: %s (err: %v)\n", argocdNamespace, output, err)

	// Wait briefly and verify the Grant still exists (protected by Usage)
	time.Sleep(10 * time.Second)

	grantName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, DatabaseGrantExpectedName, "--ignore-not-found", "-o", "jsonpath={.metadata.name}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to check Grant existence", argocdNamespace))
	require.Equal(t, DatabaseGrantExpectedName, grantName, fmt.Sprintf("[%s] Grant '%s' was deleted despite Usage protection", argocdNamespace, DatabaseGrantExpectedName))

	fmt.Printf("[%s] TEST PASSED - Usage prevented deletion of Grant '%s' while Database exists\n", argocdNamespace, DatabaseGrantExpectedName)
}
