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
	PostgresqlAdminUserKind       = "postgresqlusers.database.entigo.com"
	PostgresqlAdminUserSpecName   = "test_owner"
	PostgresqlRegularUserName     = "test-user"
	SqlGrantKind                  = "grant.postgresql.sql.m.crossplane.io"
	RegularUserExpectedGrantName  = "grant-" + PostgresqlRegularUserName + "-test-owner-" + PostgresqlInstanceName
	RegularUserExpectedUsageName  = "usage-" + RegularUserExpectedGrantName
	RegularUserExpectedSecretName = PostgresqlInstanceName + "-" + PostgresqlRegularUserName

	PostgresqlDatabaseName       = "database-one-test"
	PostgresqlDatabaseKind       = "postgresqldatabases.database.entigo.com"
	SqlDatabaseKind              = "database.postgresql.sql.m.crossplane.io"
	SqlRoleKind                  = "role.postgresql.sql.m.crossplane.io"
	SqlExtensionKind             = "extension.postgresql.sql.m.crossplane.io"
	UsageKind                    = "usage.protection.crossplane.io"
	MinimalDatabaseName          = "database-minimal-test"
	DatabaseGrantExpectedName    = PostgresqlDatabaseName + "-grant-owner-to-dbadmin"
	DatabaseTwoName              = "database-two-test"
	DatabaseTwoGrantExpectedName = DatabaseTwoName + "-grant-owner-to-dbadmin"

	AdminUserInstanceProtectionName       = PostgresqlAdminUserName + "-instance-protection"
	RegularUserInstanceProtectionName     = PostgresqlRegularUserName + "-instance-protection"
	DatabaseInstanceProtectionName        = PostgresqlDatabaseName + "-instance-protection"
	DatabaseTwoInstanceProtectionName     = DatabaseTwoName + "-instance-protection"
	MinimalDatabaseInstanceProtectionName = MinimalDatabaseName + "-instance-protection"

	DatabaseOwnerProtectionName        = PostgresqlDatabaseName + "-owner-protection"
	DatabaseTwoOwnerProtectionName     = DatabaseTwoName + "-owner-protection"
	MinimalDatabaseOwnerProtectionName = MinimalDatabaseName + "-owner-protection"
)

// runPostgresqlUserAndDatabaseTests orchestrates user and database tests.
// Admin user must be ready first, then regular user and database tests run concurrently.
// Minimal database depends on regular user, so it runs after the parallel phase.
func runPostgresqlUserAndDatabaseTests(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	waitSyncedAndReady(t, namespaceOptions, PostgresqlAdminUserKind, PostgresqlAdminUserName, 60, 10*time.Second)
	waitSyncedAndReadyByLabel(t, namespaceOptions, SqlRoleKind, PostgresqlAdminUserName, 60, 10*time.Second)
	testAdminRoleExternalNameVerified(t, namespaceOptions)
	testInstanceProtectionUsageVerified(t, namespaceOptions, AdminUserInstanceProtectionName, "Role", PostgresqlAdminUserName)

	t.Run("parallel-user-and-db", func(t *testing.T) {
		t.Run("regular-user", func(t *testing.T) {
			t.Parallel()
			waitSyncedAndReady(t, namespaceOptions, PostgresqlAdminUserKind, PostgresqlRegularUserName, 60, 10*time.Second)
			testGrantSyncedAndVerified(t, namespaceOptions, RegularUserExpectedGrantName, PostgresqlRegularUserName, PostgresqlAdminUserSpecName)
			testUsageVerified(t, namespaceOptions, RegularUserExpectedUsageName, "Role", PostgresqlRegularUserName, "Grant", RegularUserExpectedGrantName)
			testInstanceProtectionUsageVerified(t, namespaceOptions, RegularUserInstanceProtectionName, "Role", PostgresqlRegularUserName)
			testUserUsagePreventsRoleDeletion(t, namespaceOptions)
			testRegularUserExternalNameFallback(t, namespaceOptions)
			testRegularUserPrivilegesVerified(t, namespaceOptions)
			testRegularUserConnectionSecretCreated(t, namespaceOptions)
		})
		t.Run("database", func(t *testing.T) {
			t.Parallel()
			waitSyncedAndReady(t, namespaceOptions, PostgresqlDatabaseKind, PostgresqlDatabaseName, 60, 10*time.Second)
			testGrantSyncedAndVerified(t, namespaceOptions, DatabaseGrantExpectedName, "dbadmin", PostgresqlAdminUserSpecName)
			testSqlDatabaseOwnerField(t, namespaceOptions, PostgresqlDatabaseName, PostgresqlAdminUserSpecName)
			testSqlDatabaseLocaleFields(t, namespaceOptions, PostgresqlDatabaseName)
			testDatabaseExtensionsVerified(t, namespaceOptions)
			testUsageVerified(t, namespaceOptions, DatabaseOwnerProtectionName, "PostgreSQLUser", PostgresqlAdminUserName, "Database", PostgresqlDatabaseName)
			testInstanceProtectionUsageVerified(t, namespaceOptions, DatabaseInstanceProtectionName, "Database", PostgresqlDatabaseName)
		})
		t.Run("database-two", func(t *testing.T) {
			t.Parallel()
			waitSyncedAndReady(t, namespaceOptions, PostgresqlDatabaseKind, DatabaseTwoName, 60, 10*time.Second)
			testGrantSyncedAndVerified(t, namespaceOptions, DatabaseTwoGrantExpectedName, "dbadmin", PostgresqlAdminUserSpecName)
			testSqlDatabaseOwnerField(t, namespaceOptions, DatabaseTwoName, PostgresqlAdminUserSpecName)
			testSqlDatabaseLocaleFields(t, namespaceOptions, DatabaseTwoName)
			testUsageVerified(t, namespaceOptions, DatabaseTwoOwnerProtectionName, "PostgreSQLUser", PostgresqlAdminUserName, "Database", DatabaseTwoName)
			testInstanceProtectionUsageVerified(t, namespaceOptions, DatabaseTwoInstanceProtectionName, "Database", DatabaseTwoName)
		})
	})

	if t.Failed() {
		return
	}

	// Minimal database depends on regular user being ready as owner
	waitSyncedAndReady(t, namespaceOptions, PostgresqlDatabaseKind, MinimalDatabaseName, 60, 10*time.Second)
	testMinimalDatabaseDefaultsVerified(t, namespaceOptions)
	testUsageVerified(t, namespaceOptions, MinimalDatabaseOwnerProtectionName, "PostgreSQLUser", PostgresqlRegularUserName, "Database", MinimalDatabaseName)
	testInstanceProtectionUsageVerified(t, namespaceOptions, MinimalDatabaseInstanceProtectionName, "Database", MinimalDatabaseName)
	testDatabaseUsagePreventsGrantDeletion(t, namespaceOptions)
}

// testGrantSyncedAndVerified waits for a Grant to become Synced+Ready then verifies its role and memberOf fields.
func testGrantSyncedAndVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions, grantName, expectedRole, expectedMemberOf string) {
	t.Helper()
	waitSyncedAndReady(t, namespaceOptions, SqlGrantKind, grantName, 60, 10*time.Second)

	role, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, grantName, "-o", "jsonpath={.spec.forProvider.role}")
	require.NoError(t, err, "failed to get Grant role")
	require.Equal(t, expectedRole, role, "Grant '%s' role mismatch", grantName)

	memberOf, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, grantName, "-o", "jsonpath={.spec.forProvider.memberOf}")
	require.NoError(t, err, "failed to get Grant memberOf")
	require.Equal(t, expectedMemberOf, memberOf, "Grant '%s' memberOf mismatch", grantName)
}

// testSqlDatabaseOwnerField verifies the owner field of the SQL Database for a given composite.
func testSqlDatabaseOwnerField(t *testing.T, namespaceOptions *terrak8s.KubectlOptions, compositeName, expectedOwner string) {
	t.Helper()
	dbName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind,
		"-l", fmt.Sprintf("crossplane.io/composite=%s", compositeName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err, "failed to find SQL Database for composite '%s'", compositeName)
	require.NotEmpty(t, dbName, "no SQL Database found for composite '%s'", compositeName)

	owner, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.owner}")
	require.NoError(t, err, "failed to get owner")
	require.Equal(t, expectedOwner, owner, "SQL Database '%s' owner mismatch", dbName)
}

// testSqlDatabaseLocaleFields verifies encoding and locale fields of the SQL Database for a given composite.
func testSqlDatabaseLocaleFields(t *testing.T, namespaceOptions *terrak8s.KubectlOptions, compositeName string) {
	t.Helper()
	dbName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind,
		"-l", fmt.Sprintf("crossplane.io/composite=%s", compositeName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err, "failed to find SQL Database for composite '%s'", compositeName)
	require.NotEmpty(t, dbName, "no SQL Database found for composite '%s'", compositeName)

	encoding, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.encoding}")
	require.NoError(t, err, "failed to get encoding")
	require.Equal(t, "UTF8", encoding, "SQL Database '%s' encoding mismatch", dbName)

	lcCType, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.lcCType}")
	require.NoError(t, err, "failed to get lcCType")
	require.Equal(t, "et_EE.UTF-8", lcCType, "SQL Database '%s' lcCType mismatch", dbName)

	lcCollate, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.lcCollate}")
	require.NoError(t, err, "failed to get lcCollate")
	require.Equal(t, "et_EE.UTF-8", lcCollate, "SQL Database '%s' lcCollate mismatch", dbName)

	template, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.template}")
	require.NoError(t, err, "failed to get template")
	require.Equal(t, "template0", template, "SQL Database '%s' template mismatch", dbName)
}

// testUsageVerified waits for a Usage resource and verifies its of/by fields and replayDeletion.
func testUsageVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions, usageName, ofKind, ofName, byKind, byName string) {
	t.Helper()
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
	require.NoError(t, err, "Usage '%s' not found", usageName)

	actualOfKind, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, usageName, "-o", "jsonpath={.spec.of.kind}")
	require.NoError(t, err)
	require.Equal(t, ofKind, actualOfKind, "Usage '%s' spec.of.kind mismatch", usageName)

	actualOfName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, usageName, "-o", "jsonpath={.spec.of.resourceRef.name}")
	require.NoError(t, err)
	require.Equal(t, ofName, actualOfName, "Usage '%s' spec.of.resourceRef.name mismatch", usageName)

	actualByKind, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, usageName, "-o", "jsonpath={.spec.by.kind}")
	require.NoError(t, err)
	require.Equal(t, byKind, actualByKind, "Usage '%s' spec.by.kind mismatch", usageName)

	actualByName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, usageName, "-o", "jsonpath={.spec.by.resourceRef.name}")
	require.NoError(t, err)
	require.Equal(t, byName, actualByName, "Usage '%s' spec.by.resourceRef.name mismatch", usageName)

	replayDeletion, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", UsageKind, usageName, "-o", "jsonpath={.spec.replayDeletion}")
	require.NoError(t, err)
	require.Equal(t, "true", replayDeletion, "Usage '%s' replayDeletion mismatch", usageName)
}

// testInstanceProtectionUsageVerified verifies the instance-protection Usage that prevents
// PostgreSQLInstance deletion while the given user or database resource still exists.
func testInstanceProtectionUsageVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions, usageName, byKind, byName string) {
	t.Helper()
	testUsageVerified(t, namespaceOptions, usageName, "PostgreSQLInstance", PostgresqlInstanceName, byKind, byName)
}

func testAdminRoleExternalNameVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	roleName, err := getFirstByLabel(t, namespaceOptions, SqlRoleKind, PostgresqlAdminUserName)
	require.NoError(t, err, "failed to find SQL Role")
	require.NotEmpty(t, roleName, "no SQL Role found for composite '%s'", PostgresqlAdminUserName)

	externalName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", `jsonpath={.metadata.annotations.crossplane\.io/external-name}`)
	require.NoError(t, err, "failed to get crossplane.io/external-name annotation")
	require.Equal(t, PostgresqlAdminUserSpecName, externalName, "SQL Role '%s' crossplane.io/external-name expected '%s', got '%s'", roleName, PostgresqlAdminUserSpecName, externalName)
}

func testRegularUserExternalNameFallback(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	roleName, err := getFirstByLabel(t, namespaceOptions, SqlRoleKind, PostgresqlRegularUserName)
	require.NoError(t, err, "failed to find SQL Role for regular user")
	require.NotEmpty(t, roleName, "no SQL Role found for composite '%s'", PostgresqlRegularUserName)

	// When spec.name is not set, external-name should fall back to metadata.name
	externalName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", `jsonpath={.metadata.annotations.crossplane\.io/external-name}`)
	require.NoError(t, err, "failed to get crossplane.io/external-name annotation")
	require.Equal(t, PostgresqlRegularUserName, externalName, "SQL Role '%s' external-name should fall back to metadata.name '%s', got '%s'", roleName, PostgresqlRegularUserName, externalName)
}

func testRegularUserPrivilegesVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	roleName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind,
		"-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlRegularUserName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err, "failed to find SQL Role for regular user")
	require.NotEmpty(t, roleName, "no SQL Role found for composite '%s'", PostgresqlRegularUserName)

	createDb, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", "jsonpath={.spec.forProvider.privileges.createDb}")
	require.NoError(t, err, "failed to get createDb")
	require.Equal(t, "false", createDb, "SQL Role '%s' createDb mismatch", roleName)

	login, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", "jsonpath={.spec.forProvider.privileges.login}")
	require.NoError(t, err, "failed to get login")
	require.Equal(t, "true", login, "SQL Role '%s' login mismatch", roleName)

	createRole, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", "jsonpath={.spec.forProvider.privileges.createRole}")
	require.NoError(t, err, "failed to get createRole")
	require.Equal(t, "false", createRole, "SQL Role '%s' createRole mismatch", roleName)

	inherit, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", "jsonpath={.spec.forProvider.privileges.inherit}")
	require.NoError(t, err, "failed to get inherit")
	require.Equal(t, "true", inherit, "SQL Role '%s' inherit mismatch", roleName)
}

func testRegularUserConnectionSecretCreated(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for connection secret '%s'", RegularUserExpectedSecretName), 60, 10*time.Second, func() (string, error) {
		secretName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", "secret", RegularUserExpectedSecretName,
			"-n", PostgresqlNamespaceName, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if secretName == "" {
			return "", fmt.Errorf("connection secret '%s' not found", RegularUserExpectedSecretName)
		}
		return secretName, nil
	})
	require.NoError(t, err, "connection secret '%s' not found", RegularUserExpectedSecretName)
}

// testUserUsagePreventsRoleDeletion verifies that the Usage resource blocks
// deletion of the Role while the Grant still exists.
func testUserUsagePreventsRoleDeletion(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	roleName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind,
		"-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlRegularUserName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err)
	require.NotEmpty(t, roleName)

	_, _ = terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "delete", SqlRoleKind, roleName, "--wait=false")
	time.Sleep(10 * time.Second)

	existingRole, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "--ignore-not-found", "-o", "jsonpath={.metadata.name}")
	require.NoError(t, err, "failed to check Role existence")
	require.Equal(t, roleName, existingRole, "Role '%s' was deleted despite Usage protection", roleName)
}

func testMinimalDatabaseDefaultsVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	dbName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind,
		"-l", fmt.Sprintf("crossplane.io/composite=%s", MinimalDatabaseName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err, "failed to find SQL Database for minimal")
	require.NotEmpty(t, dbName, "no SQL Database found for composite '%s'", MinimalDatabaseName)

	// Owner should be the regular user (test-user) for minimal database
	owner, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlDatabaseKind, dbName, "-o", "jsonpath={.spec.forProvider.owner}")
	require.NoError(t, err, "failed to get owner")
	require.Equal(t, PostgresqlRegularUserName, owner, "SQL Database '%s' owner mismatch", dbName)
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
		require.NoError(t, err, "failed to get extension '%s'", ext.k8sName)
		require.Equal(t, ext.k8sName, extName, "extension '%s' not found", ext.k8sName)

		forProvider, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlExtensionKind, ext.k8sName, "-o", "jsonpath={.spec.forProvider.extension}")
		require.NoError(t, err, "failed to get forProvider.extension for '%s'", ext.k8sName)
		require.Equal(t, ext.extName, forProvider, "extension '%s' forProvider.extension mismatch", ext.k8sName)

		if ext.schema != "" {
			schema, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlExtensionKind, ext.k8sName, "-o", "jsonpath={.spec.forProvider.schema}")
			require.NoError(t, err, "failed to get schema for extension '%s'", ext.k8sName)
			require.Equal(t, ext.schema, schema, "extension '%s' schema mismatch", ext.k8sName)
		}
	}
}

func testDatabaseUsagePreventsGrantDeletion(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "delete", SqlGrantKind, DatabaseGrantExpectedName, "--wait=false")
	time.Sleep(10 * time.Second)

	grantName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, DatabaseGrantExpectedName, "--ignore-not-found", "-o", "jsonpath={.metadata.name}")
	require.NoError(t, err, "failed to check Grant existence")
	require.Equal(t, DatabaseGrantExpectedName, grantName, "Grant '%s' was deleted despite Usage protection", DatabaseGrantExpectedName)
}
