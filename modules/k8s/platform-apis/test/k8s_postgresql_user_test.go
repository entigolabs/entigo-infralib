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
	PostgresqlAdminUserName     = "test-admin"
	PostgresqlAdminUserKind     = "postgresqluser.database.entigo.com"
	PostgresqlAdminUserSpecName = "test_admin"
	PostgresqlRegularUserName   = "test-user"
	SqlGrantKind                = "grant.postgresql.sql.m.crossplane.io"
	// Grant name from composition: grant-{metadata.name}-{roleName | replace "_" "-"}-{instanceRef.name}
	RegularUserExpectedGrantName  = "grant-" + PostgresqlRegularUserName + "-test-admin-" + PostgresqlInstanceName
	RegularUserExpectedSecretName = PostgresqlInstanceName + "-" + PostgresqlRegularUserName
)

func runPostgresqlUserTests(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {

	testAdminUserApplied(t, argocdNamespace, namespaceOptions)
	testAdminUserSyncedAndReady(t, argocdNamespace, namespaceOptions)
	testAdminRoleSyncedAndReady(t, argocdNamespace, namespaceOptions)
	testAdminRoleExternalNameVerified(t, argocdNamespace, namespaceOptions)
	testRegularUserApplied(t, argocdNamespace, namespaceOptions)
	testRegularUserSyncedAndReady(t, argocdNamespace, namespaceOptions)
	testRegularUserGrantVerified(t, argocdNamespace, namespaceOptions)
	testRegularUserExternalNameFallback(t, argocdNamespace, namespaceOptions)
	testRegularUserPrivilegesVerified(t, argocdNamespace, namespaceOptions)
	testRegularUserConnectionSecretCreated(t, argocdNamespace, namespaceOptions)
}

func testAdminUserApplied(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Applying PostgreSQL Admin User '%s' to namespace '%s'\n", argocdNamespace, PostgresqlAdminUserName, PostgresqlNamespaceName)
	_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "apply", "-f", "./templates/postgresql_test_admin_user.yaml", "-n", PostgresqlNamespaceName)
	require.NoError(t, err, fmt.Sprintf("[%s] Applying PostgreSQL Admin User error", argocdNamespace))
	fmt.Printf("[%s] TEST PASSED - PostgreSQL Admin User applied\n", argocdNamespace)
}

func testAdminUserSyncedAndReady(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Waiting for PostgreSQL Admin User '%s' to be Synced and Ready\n", argocdNamespace, PostgresqlAdminUserName)
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
	fmt.Printf("[%s] TEST PASSED - PostgreSQL Admin User '%s' is Synced and Ready\n", argocdNamespace, PostgresqlAdminUserName)
}

func testAdminRoleSyncedAndReady(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Waiting for SQL Role related to admin user '%s' to be Synced and Ready\n", argocdNamespace, PostgresqlAdminUserName)
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
	fmt.Printf("[%s] TEST PASSED - Admin SQL Role is Synced and Ready\n", argocdNamespace)
}

func testAdminRoleExternalNameVerified(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Verifying SQL Role crossplane.io/external-name annotation for admin user '%s'\n", argocdNamespace, PostgresqlAdminUserName)
	roleName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlAdminUserName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to find SQL Role", argocdNamespace))
	require.NotEmpty(t, roleName, fmt.Sprintf("[%s] No SQL Role found for composite '%s'", argocdNamespace, PostgresqlAdminUserName))

	externalName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", `jsonpath={.metadata.annotations.crossplane\.io/external-name}`)
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get crossplane.io/external-name annotation", argocdNamespace))
	require.Equal(t, PostgresqlAdminUserSpecName, externalName, fmt.Sprintf("[%s] SQL Role '%s' crossplane.io/external-name expected '%s', got '%s'", argocdNamespace, roleName, PostgresqlAdminUserSpecName, externalName))

	fmt.Printf("[%s] TEST PASSED - Admin SQL Role crossplane.io/external-name=%s\n", argocdNamespace, PostgresqlAdminUserSpecName)
}

// ---- REGULAR USER TESTS (no spec.name, with grant) ----

func testRegularUserApplied(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Applying PostgreSQL Regular User '%s' to namespace '%s'\n", argocdNamespace, PostgresqlRegularUserName, PostgresqlNamespaceName)
	_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "apply", "-f", "./templates/postgresql_test_user.yaml", "-n", PostgresqlNamespaceName)
	require.NoError(t, err, fmt.Sprintf("[%s] Applying PostgreSQL Regular User error", argocdNamespace))
	fmt.Printf("[%s] TEST PASSED - PostgreSQL Regular User applied\n", argocdNamespace)
}

func testRegularUserSyncedAndReady(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Waiting for PostgreSQL Regular User '%s' to be Synced and Ready\n", argocdNamespace, PostgresqlRegularUserName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for PostgreSQL Regular User '%s' to be Synced and Ready", argocdNamespace, PostgresqlRegularUserName), 60, 10*time.Second, func() (string, error) {
		syncStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", PostgresqlAdminUserKind, PostgresqlRegularUserName, "-o", `jsonpath={.status.conditions[?(@.type=="Synced")].status}`)
		if err != nil {
			return "", err
		}
		if syncStatus != "True" {
			return "", fmt.Errorf("PostgreSQL Regular User '%s' not synced yet, condition: %s", PostgresqlRegularUserName, syncStatus)
		}
		readyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", PostgresqlAdminUserKind, PostgresqlRegularUserName, "-o", `jsonpath={.status.conditions[?(@.type=="Ready")].status}`)
		if err != nil {
			return "", err
		}
		if readyStatus != "True" {
			return "", fmt.Errorf("PostgreSQL Regular User '%s' not ready yet, condition: %s", PostgresqlRegularUserName, readyStatus)
		}
		return "Synced+Ready", nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] PostgreSQL Regular User '%s' failed to become Synced and Ready", argocdNamespace, PostgresqlRegularUserName))
	fmt.Printf("[%s] TEST PASSED - PostgreSQL Regular User '%s' is Synced and Ready\n", argocdNamespace, PostgresqlRegularUserName)
}

func testRegularUserGrantVerified(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Verifying Grant resource '%s' for regular user\n", argocdNamespace, RegularUserExpectedGrantName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for Grant '%s'", argocdNamespace, RegularUserExpectedGrantName), 60, 10*time.Second, func() (string, error) {
		grantName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, RegularUserExpectedGrantName, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if grantName == "" {
			return "", fmt.Errorf("Grant '%s' not found", RegularUserExpectedGrantName)
		}
		syncStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, RegularUserExpectedGrantName, "-o", `jsonpath={.status.conditions[?(@.type=="Synced")].status}`)
		if err != nil {
			return "", err
		}
		if syncStatus != "True" {
			return "", fmt.Errorf("Grant '%s' not synced yet, condition: %s", RegularUserExpectedGrantName, syncStatus)
		}
		readyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, RegularUserExpectedGrantName, "-o", `jsonpath={.status.conditions[?(@.type=="Ready")].status}`)
		if err != nil {
			return "", err
		}
		if readyStatus != "True" {
			return "", fmt.Errorf("Grant '%s' not ready yet, condition: %s", RegularUserExpectedGrantName, readyStatus)
		}
		return "Synced+Ready", nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] Grant '%s' failed to become Synced and Ready", argocdNamespace, RegularUserExpectedGrantName))

	// Verify Grant forProvider.role = metadata.name of the PostgreSQLUser
	role, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, RegularUserExpectedGrantName, "-o", "jsonpath={.spec.forProvider.role}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get Grant role", argocdNamespace))
	require.Equal(t, PostgresqlRegularUserName, role, fmt.Sprintf("[%s] Grant '%s' role mismatch", argocdNamespace, RegularUserExpectedGrantName))

	// Verify Grant forProvider.memberOf = test_admin (the PostgreSQL role name)
	memberOf, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlGrantKind, RegularUserExpectedGrantName, "-o", "jsonpath={.spec.forProvider.memberOf}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get Grant memberOf", argocdNamespace))
	require.Equal(t, PostgresqlAdminUserSpecName, memberOf, fmt.Sprintf("[%s] Grant '%s' memberOf mismatch", argocdNamespace, RegularUserExpectedGrantName))

	fmt.Printf("[%s] TEST PASSED - Grant '%s' verified (role=%s, memberOf=%s)\n", argocdNamespace, RegularUserExpectedGrantName, PostgresqlRegularUserName, PostgresqlAdminUserSpecName)
}

func testRegularUserExternalNameFallback(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Verifying SQL Role external-name fallback for regular user '%s' (spec.name not set)\n", argocdNamespace, PostgresqlRegularUserName)
	roleName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlRegularUserName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to find SQL Role for regular user", argocdNamespace))
	require.NotEmpty(t, roleName, fmt.Sprintf("[%s] No SQL Role found for composite '%s'", argocdNamespace, PostgresqlRegularUserName))

	// When spec.name is not set, external-name should fall back to metadata.name
	externalName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", `jsonpath={.metadata.annotations.crossplane\.io/external-name}`)
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get crossplane.io/external-name annotation", argocdNamespace))
	require.Equal(t, PostgresqlRegularUserName, externalName, fmt.Sprintf("[%s] SQL Role '%s' external-name should fall back to metadata.name '%s', got '%s'", argocdNamespace, roleName, PostgresqlRegularUserName, externalName))

	fmt.Printf("[%s] TEST PASSED - Regular user SQL Role external-name falls back to metadata.name=%s\n", argocdNamespace, PostgresqlRegularUserName)
}

func testRegularUserPrivilegesVerified(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Verifying SQL Role privileges for regular user '%s'\n", argocdNamespace, PostgresqlRegularUserName)
	roleName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlRegularUserName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to find SQL Role for regular user", argocdNamespace))
	require.NotEmpty(t, roleName, fmt.Sprintf("[%s] No SQL Role found for composite '%s'", argocdNamespace, PostgresqlRegularUserName))

	createDb, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", "jsonpath={.spec.forProvider.privileges.createDb}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get createDb", argocdNamespace))
	require.Equal(t, "false", createDb, fmt.Sprintf("[%s] SQL Role '%s' createDb mismatch", argocdNamespace, roleName))

	login, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", "jsonpath={.spec.forProvider.privileges.login}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get login", argocdNamespace))
	require.Equal(t, "true", login, fmt.Sprintf("[%s] SQL Role '%s' login mismatch", argocdNamespace, roleName))

	createRole, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", "jsonpath={.spec.forProvider.privileges.createRole}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get createRole", argocdNamespace))
	require.Equal(t, "false", createRole, fmt.Sprintf("[%s] SQL Role '%s' createRole mismatch", argocdNamespace, roleName))

	inherit, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlRoleKind, roleName, "-o", "jsonpath={.spec.forProvider.privileges.inherit}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get inherit", argocdNamespace))
	require.Equal(t, "true", inherit, fmt.Sprintf("[%s] SQL Role '%s' inherit mismatch", argocdNamespace, roleName))

	fmt.Printf("[%s] TEST PASSED - Regular user SQL Role privileges verified (createDb=false, login=true, createRole=false, inherit=true)\n", argocdNamespace)
}

func testRegularUserConnectionSecretCreated(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Verifying connection secret '%s' exists for regular user\n", argocdNamespace, RegularUserExpectedSecretName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for connection secret '%s'", argocdNamespace, RegularUserExpectedSecretName), 60, 10*time.Second, func() (string, error) {
		secretName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", "secret", RegularUserExpectedSecretName, "-n", PostgresqlNamespaceName, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if secretName == "" {
			return "", fmt.Errorf("connection secret '%s' not found", RegularUserExpectedSecretName)
		}
		return secretName, nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] Connection secret '%s' not found", argocdNamespace, RegularUserExpectedSecretName))
	fmt.Printf("[%s] TEST PASSED - Connection secret '%s' exists\n", argocdNamespace, RegularUserExpectedSecretName)
}
