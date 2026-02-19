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
	PostgresqlAdminUserName       = "test-admin"
	PostgresqlAdminUserKind       = "postgresqluser.database.entigo.com"
	PostgresqlAdminUserSpecName   = "test_admin"
	PostgresqlRegularUserName     = "test-user"
	SqlGrantKind                  = "grant.postgresql.sql.m.crossplane.io"
	RegularUserExpectedGrantName  = "grant-" + PostgresqlRegularUserName + "-test-admin-" + PostgresqlInstanceName
	RegularUserExpectedUsageName  = "usage-" + RegularUserExpectedGrantName
	RegularUserExpectedSecretName = PostgresqlInstanceName + "-" + PostgresqlRegularUserName

	PostgresqlDatabaseName           = "database-test"
	PostgresqlDatabaseKind           = "postgresqldatabase.database.entigo.com"
	SqlDatabaseKind                  = "database.postgresql.sql.m.crossplane.io"
	SqlRoleKind                      = "role.postgresql.sql.m.crossplane.io"
	UsageKind                        = "usage.protection.crossplane.io"
	MinimalDatabaseName              = "database-minimal-test"
	DatabaseGrantExpectedName        = "test-admin-to-dbadmin-grant"
	DatabaseUsageExpectedName        = PostgresqlDatabaseName + "-grant-usage"
	MinimalDatabaseUsageExpectedName = MinimalDatabaseName + "-grant-usage"
)

// runPostgresqlUserAndDatabaseTests orchestrates user and database tests.
// Admin user must be ready first, then regular user and database tests run concurrently.
// Minimal database depends on regular user, so it runs after the parallel phase.
func runPostgresqlUserAndDatabaseTests(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	testAdminUserApplied(t, namespaceOptions)
	testAdminUserSyncedAndReady(t, namespaceOptions)
	testAdminRoleSyncedAndReady(t, namespaceOptions)
	testAdminRoleExternalNameVerified(t, namespaceOptions)

	t.Run("parallel-user-and-db", func(t *testing.T) {
		t.Run("regular-user", func(t *testing.T) {
			t.Parallel()
			testRegularUserApplied(t, namespaceOptions)
			testRegularUserSyncedAndReady(t, namespaceOptions)
			testRegularUserGrantVerified(t, namespaceOptions)
			testRegularUserUsageVerified(t, namespaceOptions)
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
			testDatabaseUsageVerified(t, namespaceOptions)
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
	testDatabaseUsagePreventsGrantDeletion(t, namespaceOptions)
}

func testAdminUserApplied(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "apply", "-f", "./templates/postgresql_test_admin_user.yaml", "-n", PostgresqlNamespaceName)
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

	// Verify Grant forProvider.memberOf = test_admin (the PostgreSQL role name)
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
	_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "apply", "-f", "./templates/postgresql_test_database.yaml", "-n", PostgresqlNamespaceName)
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
	// Minimal database owner is "test-user", so grant name is "test-user-to-dbadmin-grant"
	expectedGrantName := "test-user-to-dbadmin-grant"

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
