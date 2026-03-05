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
	PostgresqlAdminUserName       = "test-owner"
	PostgresqlAdminUserKind       = "postgresqluser.database.entigo.com"
	PostgresqlAdminUserSpecName   = "test_owner"
	PostgresqlRegularUserName     = "test-user"
	SqlGrantKind                  = "grant.postgresql.sql.m.crossplane.io"
	RegularUserExpectedGrantName  = "grant-" + PostgresqlRegularUserName + "-test-owner-" + PostgresqlInstanceName
	RegularUserExpectedUsageName  = "usage-" + RegularUserExpectedGrantName
	RegularUserExpectedSecretName = PostgresqlInstanceName + "-" + PostgresqlRegularUserName

	PostgresqlDatabaseName           = "database-one-test"
	PostgresqlDatabaseKind           = "postgresqldatabase.database.entigo.com"
	SqlDatabaseKind                  = "database.postgresql.sql.m.crossplane.io"
	SqlRoleKind                      = "role.postgresql.sql.m.crossplane.io"
	SqlExtensionKind                 = "extension.postgresql.sql.m.crossplane.io"
	UsageKind                        = "usage.protection.crossplane.io"
	MinimalDatabaseName              = "database-minimal-test"
	DatabaseGrantExpectedName        = PostgresqlDatabaseName + "-grant-owner-to-dbadmin"
	DatabaseUsageExpectedName        = PostgresqlDatabaseName + "-grant-usage"
	MinimalDatabaseUsageExpectedName = MinimalDatabaseName + "-grant-usage"
	DatabaseTwoName                  = "database-two-test"
	DatabaseTwoGrantExpectedName     = DatabaseTwoName + "-grant-owner-to-dbadmin"
	DatabaseTwoUsageExpectedName     = DatabaseTwoName + "-grant-usage"

	AdminUserInstanceProtectionName       = PostgresqlAdminUserName + "-instance-protection"
	RegularUserInstanceProtectionName     = PostgresqlRegularUserName + "-instance-protection"
	DatabaseInstanceProtectionName        = PostgresqlDatabaseName + "-instance-protection"
	DatabaseTwoInstanceProtectionName     = DatabaseTwoName + "-instance-protection"
	MinimalDatabaseInstanceProtectionName = MinimalDatabaseName + "-instance-protection"
)

// runPostgresqlUserAndDatabaseTests orchestrates user and database tests.
// Admin user must be ready first, then regular user and database tests run concurrently.
// Minimal database depends on regular user, so it runs after the parallel phase.
func runPostgresqlUserAndDatabaseTests(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	testAdminUserApplied(t, namespaceOptions)
	testAdminUserSyncedAndReady(t, namespaceOptions)
	testAdminRoleSyncedAndReady(t, namespaceOptions)
	testAdminRoleExternalNameVerified(t, namespaceOptions)
	testSequenceRoleCreatedAfterInstanceReady(t, namespaceOptions, PostgresqlAdminUserName)
	testInstanceProtectionUsageVerified(t, namespaceOptions, AdminUserInstanceProtectionName, "PostgreSQLUser", PostgresqlAdminUserName)

	t.Run("parallel-user-and-db", func(t *testing.T) {
		t.Run("regular-user", func(t *testing.T) {
			t.Parallel()
			testRegularUserApplied(t, namespaceOptions)
			testRegularUserSyncedAndReady(t, namespaceOptions)
			testRegularUserGrantVerified(t, namespaceOptions)
			testRegularUserUsageVerified(t, namespaceOptions)
			testInstanceProtectionUsageVerified(t, namespaceOptions, RegularUserInstanceProtectionName, "PostgreSQLUser", PostgresqlRegularUserName)
			testUserUsagePreventsRoleDeletion(t, namespaceOptions)
			testRegularUserExternalNameFallback(t, namespaceOptions)
			testRegularUserPrivilegesVerified(t, namespaceOptions)
			testRegularUserConnectionSecretCreated(t, namespaceOptions)
		})
		t.Run("database", func(t *testing.T) {
			t.Parallel()
			testDatabaseApplied(t, namespaceOptions)
			testDatabaseSyncedAndReady(t, namespaceOptions)
			testDatabaseGrantOwnerToDbadmin(t, namespaceOptions)
			testDatabaseOwnerFieldVerified(t, namespaceOptions)
			testDatabaseFieldsVerified(t, namespaceOptions)
			testDatabaseExtensionsVerified(t, namespaceOptions)
			testDatabaseUsageVerified(t, namespaceOptions)
			testInstanceProtectionUsageVerified(t, namespaceOptions, DatabaseInstanceProtectionName, "PostgreSQLDatabase", PostgresqlDatabaseName)
		})
		t.Run("database-two", func(t *testing.T) {
			t.Parallel()
			testDatabaseTwoApplied(t, namespaceOptions)
			testDatabaseTwoSyncedAndReady(t, namespaceOptions)
			testDatabaseTwoGrantOwnerToDbadmin(t, namespaceOptions)
			testDatabaseTwoOwnerFieldVerified(t, namespaceOptions)
			testDatabaseTwoFieldsVerified(t, namespaceOptions)
			testDatabaseTwoUsageVerified(t, namespaceOptions)
			testInstanceProtectionUsageVerified(t, namespaceOptions, DatabaseTwoInstanceProtectionName, "PostgreSQLDatabase", DatabaseTwoName)
		})
	})

	if t.Failed() {
		return
	}

	// Minimal database depends on regular user being ready as owner
	testMinimalDatabaseApplied(t, namespaceOptions)
	testMinimalDatabaseSyncedAndReady(t, namespaceOptions)
	testMinimalDatabaseDefaultsVerified(t, namespaceOptions)
	testMinimalDatabaseUsageVerified(t, namespaceOptions)
	testInstanceProtectionUsageVerified(t, namespaceOptions, MinimalDatabaseInstanceProtectionName, "PostgreSQLDatabase", MinimalDatabaseName)
	testSequenceMinimalDatabaseGrantAfterUserReady(t, namespaceOptions)
	testDatabaseUsagePreventsGrantDeletion(t, namespaceOptions)
}

func testAdminUserApplied(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "apply", "-f", "./templates/postgresql_test_owner_user.yaml", "-n", PostgresqlNamespaceName)
	require.NoError(t, err, "applying PostgreSQL Admin User error")
}

func testAdminUserSyncedAndReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	waitSyncedAndReady(t, namespaceOptions, PostgresqlAdminUserKind, PostgresqlAdminUserName, 60, 10*time.Second)
}

func testAdminRoleSyncedAndReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	waitSyncedAndReadyByLabel(t, namespaceOptions, SqlRoleKind, PostgresqlAdminUserName, 60, 10*time.Second)
}

func testAdminRoleExternalNameVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	roleName, err := getFirstByLabel(t, namespaceOptions, SqlRoleKind, PostgresqlAdminUserName)
	require.NoError(t, err, "failed to find SQL Role")
	require.NotEmpty(t, roleName, fmt.Sprintf("no SQL Role found for composite '%s'", PostgresqlAdminUserName))

	externalName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", `jsonpath={.metadata.annotations.crossplane\.io/external-name}`)
	require.NoError(t, err, "failed to get crossplane.io/external-name annotation")
	require.Equal(t, PostgresqlAdminUserSpecName, externalName, fmt.Sprintf("SQL Role '%s' crossplane.io/external-name expected '%s', got '%s'", roleName, PostgresqlAdminUserSpecName, externalName))
}

func testRegularUserApplied(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "apply", "-f", "./templates/postgresql_test_user.yaml", "-n", PostgresqlNamespaceName)
	require.NoError(t, err, "applying PostgreSQL Regular User error")
}

func testRegularUserSyncedAndReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	waitSyncedAndReady(t, namespaceOptions, PostgresqlAdminUserKind, PostgresqlRegularUserName, 60, 10*time.Second)
}

func testRegularUserGrantVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for Grant '%s'", RegularUserExpectedGrantName), 60, 10*time.Second, func() (string, error) {
		grantName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, RegularUserExpectedGrantName, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if grantName == "" {
			return "", fmt.Errorf("Grant '%s' not found", RegularUserExpectedGrantName)
		}
		for _, condType := range []string{"Synced", "Ready"} {
			status, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, RegularUserExpectedGrantName, "-o",
				fmt.Sprintf(`jsonpath={.status.conditions[?(@.type=="%s")].status}`, condType))
			if err != nil {
				return "", err
			}
			if status != "True" {
				return "", fmt.Errorf("Grant '%s': %s=%s", RegularUserExpectedGrantName, condType, status)
			}
		}
		return "Synced+Ready", nil
	})
	require.NoError(t, err, fmt.Sprintf("Grant '%s' failed to become Synced and Ready", RegularUserExpectedGrantName))

	// Verify Grant forProvider.role = metadata.name of the PostgreSQLUser
	role, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, RegularUserExpectedGrantName, "-o", "jsonpath={.spec.forProvider.role}")
	require.NoError(t, err, "failed to get Grant role")
	require.Equal(t, PostgresqlRegularUserName, role, fmt.Sprintf("Grant '%s' role mismatch", RegularUserExpectedGrantName))

	// Verify Grant forProvider.memberOf = test_owner (the PostgreSQL role name)
	memberOf, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, RegularUserExpectedGrantName, "-o", "jsonpath={.spec.forProvider.memberOf}")
	require.NoError(t, err, "failed to get Grant memberOf")
	require.Equal(t, PostgresqlAdminUserSpecName, memberOf, fmt.Sprintf("Grant '%s' memberOf mismatch", RegularUserExpectedGrantName))
}

func testRegularUserExternalNameFallback(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	roleName, err := getFirstByLabel(t, namespaceOptions, SqlRoleKind, PostgresqlRegularUserName)
	require.NoError(t, err, "failed to find SQL Role for regular user")
	require.NotEmpty(t, roleName, fmt.Sprintf("no SQL Role found for composite '%s'", PostgresqlRegularUserName))

	// When spec.name is not set, external-name should fall back to metadata.name
	externalName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", `jsonpath={.metadata.annotations.crossplane\.io/external-name}`)
	require.NoError(t, err, "failed to get crossplane.io/external-name annotation")
	require.Equal(t, PostgresqlRegularUserName, externalName, fmt.Sprintf("SQL Role '%s' external-name should fall back to metadata.name '%s', got '%s'", roleName, PostgresqlRegularUserName, externalName))
}

func testRegularUserPrivilegesVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	roleName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlRegularUserName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err, "failed to find SQL Role for regular user")
	require.NotEmpty(t, roleName, fmt.Sprintf("no SQL Role found for composite '%s'", PostgresqlRegularUserName))

	createDb, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", "jsonpath={.spec.forProvider.privileges.createDb}")
	require.NoError(t, err, "failed to get createDb")
	require.Equal(t, "false", createDb, fmt.Sprintf("SQL Role '%s' createDb mismatch", roleName))

	login, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", "jsonpath={.spec.forProvider.privileges.login}")
	require.NoError(t, err, "failed to get login")
	require.Equal(t, "true", login, fmt.Sprintf("SQL Role '%s' login mismatch", roleName))

	createRole, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", "jsonpath={.spec.forProvider.privileges.createRole}")
	require.NoError(t, err, "failed to get createRole")
	require.Equal(t, "false", createRole, fmt.Sprintf("SQL Role '%s' createRole mismatch", roleName))

	inherit, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", "jsonpath={.spec.forProvider.privileges.inherit}")
	require.NoError(t, err, "failed to get inherit")
	require.Equal(t, "true", inherit, fmt.Sprintf("SQL Role '%s' inherit mismatch", roleName))
}

func testRegularUserConnectionSecretCreated(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for connection secret '%s'", RegularUserExpectedSecretName), 60, 10*time.Second, func() (string, error) {
		secretName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", "secret", RegularUserExpectedSecretName, "-n", PostgresqlNamespaceName, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if secretName == "" {
			return "", fmt.Errorf("connection secret '%s' not found", RegularUserExpectedSecretName)
		}
		return secretName, nil
	})
	require.NoError(t, err, fmt.Sprintf("connection secret '%s' not found", RegularUserExpectedSecretName))
}

func testRegularUserUsageVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for Usage '%s'", RegularUserExpectedUsageName), 30, 10*time.Second, func() (string, error) {
		name, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, RegularUserExpectedUsageName, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if name == "" {
			return "", fmt.Errorf("Usage '%s' not found", RegularUserExpectedUsageName)
		}
		return name, nil
	})
	require.NoError(t, err, fmt.Sprintf("Usage '%s' not found", RegularUserExpectedUsageName))

	ofKind, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, RegularUserExpectedUsageName, "-o", "jsonpath={.spec.of.kind}")
	require.NoError(t, err)
	require.Equal(t, "Role", ofKind)

	ofName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, RegularUserExpectedUsageName, "-o", "jsonpath={.spec.of.resourceRef.name}")
	require.NoError(t, err)
	require.Equal(t, PostgresqlRegularUserName, ofName)

	byKind, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, RegularUserExpectedUsageName, "-o", "jsonpath={.spec.by.kind}")
	require.NoError(t, err)
	require.Equal(t, "Grant", byKind)

	byName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, RegularUserExpectedUsageName, "-o", "jsonpath={.spec.by.resourceRef.name}")
	require.NoError(t, err)
	require.Equal(t, RegularUserExpectedGrantName, byName)

	replayDeletion, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, RegularUserExpectedUsageName, "-o", "jsonpath={.spec.replayDeletion}")
	require.NoError(t, err)
	require.Equal(t, "true", replayDeletion)
}

// testUserUsagePreventsRoleDeletion verifies that the Usage resource blocks
// deletion of the Role while the Grant still exists.
func testUserUsagePreventsRoleDeletion(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	roleName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlRegularUserName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err)
	require.NotEmpty(t, roleName)

	_, _ = terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "delete", SqlRoleKind, roleName, "--wait=false")
	time.Sleep(10 * time.Second)

	existingRole, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "--ignore-not-found", "-o", "jsonpath={.metadata.name}")
	require.NoError(t, err, "failed to check Role existence")
	require.Equal(t, roleName, existingRole, fmt.Sprintf("Role '%s' was deleted despite Usage protection", roleName))
}

func testDatabaseApplied(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "apply", "-f", "./templates/postgresql_test_database_one.yaml", "-n", PostgresqlNamespaceName)
	require.NoError(t, err, "applying PostgreSQL Database error")
}

func testDatabaseSyncedAndReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	waitSyncedAndReady(t, namespaceOptions, PostgresqlDatabaseKind, PostgresqlDatabaseName, 60, 10*time.Second)
}

func testDatabaseGrantOwnerToDbadmin(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for Grant '%s'", DatabaseGrantExpectedName), 60, 10*time.Second, func() (string, error) {
		grantName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, DatabaseGrantExpectedName, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if grantName == "" {
			return "", fmt.Errorf("Grant '%s' not found", DatabaseGrantExpectedName)
		}
		for _, condType := range []string{"Synced", "Ready"} {
			status, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, DatabaseGrantExpectedName, "-o",
				fmt.Sprintf(`jsonpath={.status.conditions[?(@.type=="%s")].status}`, condType))
			if err != nil {
				return "", err
			}
			if status != "True" {
				return "", fmt.Errorf("Grant '%s': %s=%s", DatabaseGrantExpectedName, condType, status)
			}
		}
		return "Synced+Ready", nil
	})
	require.NoError(t, err, fmt.Sprintf("Grant '%s' failed to become Synced and Ready", DatabaseGrantExpectedName))

	role, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, DatabaseGrantExpectedName, "-o", "jsonpath={.spec.forProvider.role}")
	require.NoError(t, err, "failed to get Grant role")
	require.Equal(t, "dbadmin", role, fmt.Sprintf("Grant '%s' role mismatch", DatabaseGrantExpectedName))

	memberOf, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, DatabaseGrantExpectedName, "-o", "jsonpath={.spec.forProvider.memberOf}")
	require.NoError(t, err, "failed to get Grant memberOf")
	require.Equal(t, PostgresqlAdminUserSpecName, memberOf, fmt.Sprintf("Grant '%s' memberOf mismatch", DatabaseGrantExpectedName))
}

func testDatabaseOwnerFieldVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	dbName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlDatabaseName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err, "failed to find SQL Database")
	require.NotEmpty(t, dbName, fmt.Sprintf("no SQL Database found for composite '%s'", PostgresqlDatabaseName))

	owner, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.owner}")
	require.NoError(t, err, "failed to get owner")
	require.Equal(t, PostgresqlAdminUserSpecName, owner, fmt.Sprintf("SQL Database '%s' owner mismatch", dbName))
}

func testDatabaseFieldsVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	dbName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlDatabaseName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err, "failed to find SQL Database")
	require.NotEmpty(t, dbName, fmt.Sprintf("no SQL Database found for composite '%s'", PostgresqlDatabaseName))

	encoding, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.encoding}")
	require.NoError(t, err, "failed to get encoding")
	require.Equal(t, "UTF8", encoding, fmt.Sprintf("SQL Database '%s' encoding mismatch", dbName))

	lcCType, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.lcCType}")
	require.NoError(t, err, "failed to get lcCType")
	require.Equal(t, "et_EE.UTF-8", lcCType, fmt.Sprintf("SQL Database '%s' lcCType mismatch", dbName))

	lcCollate, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.lcCollate}")
	require.NoError(t, err, "failed to get lcCollate")
	require.Equal(t, "et_EE.UTF-8", lcCollate, fmt.Sprintf("SQL Database '%s' lcCollate mismatch", dbName))

	template, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.template}")
	require.NoError(t, err, "failed to get template")
	require.Equal(t, "template0", template, fmt.Sprintf("SQL Database '%s' template mismatch", dbName))
}

func testMinimalDatabaseApplied(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "apply", "-f", "./templates/postgresql_test_database_minimal.yaml", "-n", PostgresqlNamespaceName)
	require.NoError(t, err, "applying minimal PostgreSQL Database error")
}

func testMinimalDatabaseSyncedAndReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	waitSyncedAndReady(t, namespaceOptions, PostgresqlDatabaseKind, MinimalDatabaseName, 60, 10*time.Second)
}

func testMinimalDatabaseDefaultsVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	dbName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", MinimalDatabaseName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err, "failed to find SQL Database for minimal")
	require.NotEmpty(t, dbName, fmt.Sprintf("no SQL Database found for composite '%s'", MinimalDatabaseName))

	// Owner should still be set (test-user for minimal database)
	owner, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.owner}")
	require.NoError(t, err, "failed to get owner")
	require.Equal(t, PostgresqlRegularUserName, owner, fmt.Sprintf("SQL Database '%s' owner mismatch", dbName))
}

func testDatabaseUsageVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for Usage '%s'", DatabaseUsageExpectedName), 30, 10*time.Second, func() (string, error) {
		name, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, DatabaseUsageExpectedName, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if name == "" {
			return "", fmt.Errorf("Usage '%s' not found", DatabaseUsageExpectedName)
		}
		return name, nil
	})
	require.NoError(t, err, fmt.Sprintf("Usage '%s' not found", DatabaseUsageExpectedName))

	ofKind, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, DatabaseUsageExpectedName, "-o", "jsonpath={.spec.of.kind}")
	require.NoError(t, err)
	require.Equal(t, "Grant", ofKind)

	ofName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, DatabaseUsageExpectedName, "-o", "jsonpath={.spec.of.resourceRef.name}")
	require.NoError(t, err)
	require.Equal(t, DatabaseGrantExpectedName, ofName)

	byKind, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, DatabaseUsageExpectedName, "-o", "jsonpath={.spec.by.kind}")
	require.NoError(t, err)
	require.Equal(t, "Database", byKind)

	byName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, DatabaseUsageExpectedName, "-o", "jsonpath={.spec.by.resourceRef.name}")
	require.NoError(t, err)
	require.Equal(t, PostgresqlDatabaseName, byName)

	replayDeletion, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, DatabaseUsageExpectedName, "-o", "jsonpath={.spec.replayDeletion}")
	require.NoError(t, err)
	require.Equal(t, "true", replayDeletion)
}

func testMinimalDatabaseUsageVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	expectedGrantName := MinimalDatabaseName + "-grant-owner-to-dbadmin"

	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for Usage '%s'", MinimalDatabaseUsageExpectedName), 30, 10*time.Second, func() (string, error) {
		name, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, MinimalDatabaseUsageExpectedName, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if name == "" {
			return "", fmt.Errorf("Usage '%s' not found", MinimalDatabaseUsageExpectedName)
		}
		return name, nil
	})
	require.NoError(t, err, fmt.Sprintf("Usage '%s' not found", MinimalDatabaseUsageExpectedName))

	ofName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, MinimalDatabaseUsageExpectedName, "-o", "jsonpath={.spec.of.resourceRef.name}")
	require.NoError(t, err)
	require.Equal(t, expectedGrantName, ofName)

	byName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, MinimalDatabaseUsageExpectedName, "-o", "jsonpath={.spec.by.resourceRef.name}")
	require.NoError(t, err)
	require.Equal(t, MinimalDatabaseName, byName)
}

func testDatabaseUsagePreventsGrantDeletion(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "delete", SqlGrantKind, DatabaseGrantExpectedName, "--wait=false")
	time.Sleep(10 * time.Second)

	grantName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, DatabaseGrantExpectedName, "--ignore-not-found", "-o", "jsonpath={.metadata.name}")
	require.NoError(t, err, "failed to check Grant existence")
	require.Equal(t, DatabaseGrantExpectedName, grantName, fmt.Sprintf("Grant '%s' was deleted despite Usage protection", DatabaseGrantExpectedName))
}

func testDatabaseExtensionsVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	type extExpectation struct {
		k8sName string
		extName string
		schema  string
	}
	expected := []extExpectation{
		{PostgresqlDatabaseName + "-postgis", "postgis", ""},
		{PostgresqlDatabaseName + "-postgis-topology", "postgis_topology", "topology"},
		{PostgresqlDatabaseName + "-fuzzystrmatch", "fuzzystrmatch", ""},
		{PostgresqlDatabaseName + "-postgis-tiger-geocoder", "postgis_tiger_geocoder", "tiger"},
		{PostgresqlDatabaseName + "-uuid-ossp", "uuid-ossp", ""},
		{PostgresqlDatabaseName + "-btree-gist", "btree_gist", ""},
	}

	for _, ext := range expected {
		ext := ext
		extName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlExtensionKind, ext.k8sName, "--ignore-not-found", "-o", "jsonpath={.metadata.name}")
		require.NoError(t, err, fmt.Sprintf("failed to get extension '%s'", ext.k8sName))
		require.Equal(t, ext.k8sName, extName, fmt.Sprintf("extension '%s' not found", ext.k8sName))

		forProvider, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlExtensionKind, ext.k8sName, "-o", "jsonpath={.spec.forProvider.extension}")
		require.NoError(t, err, fmt.Sprintf("failed to get forProvider.extension for '%s'", ext.k8sName))
		require.Equal(t, ext.extName, forProvider, fmt.Sprintf("extension '%s' forProvider.extension mismatch", ext.k8sName))

		if ext.schema != "" {
			schema, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlExtensionKind, ext.k8sName, "-o", "jsonpath={.spec.forProvider.schema}")
			require.NoError(t, err, fmt.Sprintf("failed to get schema for extension '%s'", ext.k8sName))
			require.Equal(t, ext.schema, schema, fmt.Sprintf("extension '%s' schema mismatch", ext.k8sName))
		}
	}
}

func testDatabaseTwoApplied(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "apply", "-f", "./templates/postgresql_test_database_two.yaml", "-n", PostgresqlNamespaceName)
	require.NoError(t, err, "applying PostgreSQL Database Two error")
}

func testDatabaseTwoSyncedAndReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	waitSyncedAndReady(t, namespaceOptions, PostgresqlDatabaseKind, DatabaseTwoName, 60, 10*time.Second)
}

func testDatabaseTwoGrantOwnerToDbadmin(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for Grant '%s'", DatabaseTwoGrantExpectedName), 60, 10*time.Second, func() (string, error) {
		grantName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, DatabaseTwoGrantExpectedName, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if grantName == "" {
			return "", fmt.Errorf("Grant '%s' not found", DatabaseTwoGrantExpectedName)
		}
		for _, condType := range []string{"Synced", "Ready"} {
			status, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, DatabaseTwoGrantExpectedName, "-o",
				fmt.Sprintf(`jsonpath={.status.conditions[?(@.type=="%s")].status}`, condType))
			if err != nil {
				return "", err
			}
			if status != "True" {
				return "", fmt.Errorf("Grant '%s': %s=%s", DatabaseTwoGrantExpectedName, condType, status)
			}
		}
		return "Synced+Ready", nil
	})
	require.NoError(t, err, fmt.Sprintf("Grant '%s' failed to become Synced and Ready", DatabaseTwoGrantExpectedName))

	role, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, DatabaseTwoGrantExpectedName, "-o", "jsonpath={.spec.forProvider.role}")
	require.NoError(t, err, "failed to get Grant role")
	require.Equal(t, "dbadmin", role, fmt.Sprintf("Grant '%s' role mismatch", DatabaseTwoGrantExpectedName))

	memberOf, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, DatabaseTwoGrantExpectedName, "-o", "jsonpath={.spec.forProvider.memberOf}")
	require.NoError(t, err, "failed to get Grant memberOf")
	require.Equal(t, PostgresqlAdminUserSpecName, memberOf, fmt.Sprintf("Grant '%s' memberOf mismatch", DatabaseTwoGrantExpectedName))
}

func testDatabaseTwoOwnerFieldVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	dbName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", DatabaseTwoName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err, "failed to find SQL Database Two")
	require.NotEmpty(t, dbName, fmt.Sprintf("no SQL Database found for composite '%s'", DatabaseTwoName))

	owner, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.owner}")
	require.NoError(t, err, "failed to get owner")
	require.Equal(t, PostgresqlAdminUserSpecName, owner, fmt.Sprintf("SQL Database '%s' owner mismatch", dbName))
}

func testDatabaseTwoFieldsVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	dbName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", DatabaseTwoName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err, "failed to find SQL Database Two")
	require.NotEmpty(t, dbName, fmt.Sprintf("no SQL Database found for composite '%s'", DatabaseTwoName))

	encoding, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.encoding}")
	require.NoError(t, err, "failed to get encoding")
	require.Equal(t, "UTF8", encoding, fmt.Sprintf("SQL Database '%s' encoding mismatch", dbName))

	lcCType, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.lcCType}")
	require.NoError(t, err, "failed to get lcCType")
	require.Equal(t, "et_EE.UTF-8", lcCType, fmt.Sprintf("SQL Database '%s' lcCType mismatch", dbName))

	lcCollate, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.lcCollate}")
	require.NoError(t, err, "failed to get lcCollate")
	require.Equal(t, "et_EE.UTF-8", lcCollate, fmt.Sprintf("SQL Database '%s' lcCollate mismatch", dbName))

	template, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.template}")
	require.NoError(t, err, "failed to get template")
	require.Equal(t, "template0", template, fmt.Sprintf("SQL Database '%s' template mismatch", dbName))
}

func testDatabaseTwoUsageVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for Usage '%s'", DatabaseTwoUsageExpectedName), 30, 10*time.Second, func() (string, error) {
		name, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, DatabaseTwoUsageExpectedName, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if name == "" {
			return "", fmt.Errorf("Usage '%s' not found", DatabaseTwoUsageExpectedName)
		}
		return name, nil
	})
	require.NoError(t, err, fmt.Sprintf("Usage '%s' not found", DatabaseTwoUsageExpectedName))

	ofKind, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, DatabaseTwoUsageExpectedName, "-o", "jsonpath={.spec.of.kind}")
	require.NoError(t, err)
	require.Equal(t, "Grant", ofKind)

	ofName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, DatabaseTwoUsageExpectedName, "-o", "jsonpath={.spec.of.resourceRef.name}")
	require.NoError(t, err)
	require.Equal(t, DatabaseTwoGrantExpectedName, ofName)

	byKind, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, DatabaseTwoUsageExpectedName, "-o", "jsonpath={.spec.by.kind}")
	require.NoError(t, err)
	require.Equal(t, "Database", byKind)

	byName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, DatabaseTwoUsageExpectedName, "-o", "jsonpath={.spec.by.resourceRef.name}")
	require.NoError(t, err)
	require.Equal(t, DatabaseTwoName, byName)

	replayDeletion, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, DatabaseTwoUsageExpectedName, "-o", "jsonpath={.spec.replayDeletion}")
	require.NoError(t, err)
	require.Equal(t, "true", replayDeletion)
}

// testInstanceProtectionUsageVerified verifies the instance-protection Usage that prevents
// PostgreSQLInstance deletion while the given user or database resource still exists.
func testInstanceProtectionUsageVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions, usageName string, byKind string, byName string) {
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for Usage '%s'", usageName), 30, 10*time.Second, func() (string, error) {
		name, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, usageName, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if name == "" {
			return "", fmt.Errorf("Usage '%s' not found", usageName)
		}
		return name, nil
	})
	require.NoError(t, err, fmt.Sprintf("Usage '%s' not found", usageName))

	ofKind, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, usageName, "-o", "jsonpath={.spec.of.kind}")
	require.NoError(t, err)
	require.Equal(t, "PostgreSQLInstance", ofKind)

	ofName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, usageName, "-o", "jsonpath={.spec.of.resourceRef.name}")
	require.NoError(t, err)
	require.Equal(t, PostgresqlInstanceName, ofName)

	byKindActual, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, usageName, "-o", "jsonpath={.spec.by.kind}")
	require.NoError(t, err)
	require.Equal(t, byKind, byKindActual)

	byNameActual, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, usageName, "-o", "jsonpath={.spec.by.resourceRef.name}")
	require.NoError(t, err)
	require.Equal(t, byName, byNameActual)

	replayDeletion, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, usageName, "-o", "jsonpath={.spec.replayDeletion}")
	require.NoError(t, err)
	require.Equal(t, "true", replayDeletion)
}

// testSequenceRoleCreatedAfterInstanceReady verifies that the SQL Role for the given user
// was created only after the PostgreSQLInstance became Ready — proving the cross-XR creation gate works.
func testSequenceRoleCreatedAfterInstanceReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions, compositeUserName string) {
	instanceReadyTimeStr, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", PostgresqlInstanceKind, PostgresqlInstanceName,
		"-o", `jsonpath={.status.conditions[?(@.type=="Ready")].lastTransitionTime}`)
	require.NoError(t, err, "failed to get PostgreSQLInstance Ready transition time")
	require.NotEmpty(t, instanceReadyTimeStr, "PostgreSQLInstance Ready condition not found")

	roleName, err := getFirstByLabel(t, namespaceOptions, SqlRoleKind, compositeUserName)
	require.NoError(t, err, "failed to find SQL Role for user")
	require.NotEmpty(t, roleName, fmt.Sprintf("no SQL Role found for composite '%s'", compositeUserName))

	roleCreationTimeStr, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName,
		"-o", "jsonpath={.metadata.creationTimestamp}")
	require.NoError(t, err, "failed to get SQL Role creation time")

	instanceReadyTime, err := time.Parse(time.RFC3339, instanceReadyTimeStr)
	require.NoError(t, err, "failed to parse instance ready time")
	roleCreationTime, err := time.Parse(time.RFC3339, roleCreationTimeStr)
	require.NoError(t, err, "failed to parse role creation time")

	require.False(t, roleCreationTime.Before(instanceReadyTime),
		fmt.Sprintf("SQL Role for '%s' was created at %s before PostgreSQLInstance became Ready at %s — cross-XR sequence gate failed",
			compositeUserName, roleCreationTime, instanceReadyTime))
}

// testSequenceMinimalDatabaseGrantAfterUserReady verifies that the minimal database's Grant
// was created only after the regular user's Role — proving the user-gates-database cross-XR sequence works.
func testSequenceMinimalDatabaseGrantAfterUserReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	regularUserRoleName, err := getFirstByLabel(t, namespaceOptions, SqlRoleKind, PostgresqlRegularUserName)
	require.NoError(t, err, "failed to find SQL Role for regular user")
	require.NotEmpty(t, regularUserRoleName, fmt.Sprintf("no SQL Role found for composite '%s'", PostgresqlRegularUserName))

	roleCreationTimeStr, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, regularUserRoleName,
		"-o", "jsonpath={.metadata.creationTimestamp}")
	require.NoError(t, err, "failed to get regular user Role creation time")

	minimalGrantName := MinimalDatabaseName + "-grant-owner-to-dbadmin"
	grantCreationTimeStr, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, minimalGrantName,
		"-o", "jsonpath={.metadata.creationTimestamp}")
	require.NoError(t, err, "failed to get minimal database Grant creation time")
	require.NotEmpty(t, grantCreationTimeStr, fmt.Sprintf("Grant '%s' not found", minimalGrantName))

	roleCreationTime, err := time.Parse(time.RFC3339, roleCreationTimeStr)
	require.NoError(t, err, "failed to parse role creation time")
	grantCreationTime, err := time.Parse(time.RFC3339, grantCreationTimeStr)
	require.NoError(t, err, "failed to parse grant creation time")

	require.False(t, grantCreationTime.Before(roleCreationTime),
		fmt.Sprintf("minimal database Grant created at %s before regular user Role at %s — user-gates-database sequence failed",
			grantCreationTime, roleCreationTime))
}
