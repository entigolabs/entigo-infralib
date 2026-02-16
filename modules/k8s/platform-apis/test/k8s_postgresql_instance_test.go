package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
)

const (
	PostgresqlInstanceName     = "postgresql-instance-test"
	PostgresqlInstanceKind     = "postgresqlinstance.database.entigo.com"
	RdsInstanceKind            = "instance.rds.aws.m.upbound.io"
	SecurityGroupKind          = "securitygroup.ec2.aws.m.upbound.io"
	SecurityGroupRuleKind      = "securitygrouprule.ec2.aws.m.upbound.io"
	ExternalSecretKind         = "externalsecret.external-secrets.io"
	SqlProviderConfigKind      = "providerconfig.postgresql.sql.m.crossplane.io"
	ProviderConfigExpectedName = PostgresqlInstanceName + "-providerconfig"
)

func runPostgresqlInstanceTests(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	testPostgresqlInstanceApplied(t, argocdNamespace, namespaceOptions)
	testPostgresqlInstanceSyncedAndReady(t, argocdNamespace, namespaceOptions)
	testSecurityGroupRulesSyncedAndReady(t, argocdNamespace, namespaceOptions)
	testSecurityGroupSyncedAndReady(t, argocdNamespace, namespaceOptions)
	testExternalSecretReady(t, argocdNamespace, namespaceOptions)
	testProviderConfigExists(t, argocdNamespace, namespaceOptions)
	testRdsInstanceSyncedAndReady(t, argocdNamespace, namespaceOptions)
	testRdsInstanceFieldsVerified(t, argocdNamespace, namespaceOptions)
	testDeletionProtectionToggle(t, argocdNamespace, namespaceOptions)
}

func testPostgresqlInstanceApplied(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Applying PostgreSQL Instance '%s' to namespace '%s'\n", argocdNamespace, PostgresqlInstanceName, PostgresqlNamespaceName)
	_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "apply", "-f", "./templates/postgresql_test_instance.yaml", "-n", PostgresqlNamespaceName)
	require.NoError(t, err, fmt.Sprintf("[%s] Applying PostgreSQL error", argocdNamespace))
	fmt.Printf("[%s] TEST PASSED - PostgreSQL applied with deletionProtection=false\n", argocdNamespace)
}

func testPostgresqlInstanceSyncedAndReady(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Waiting for PostgreSQL Instance '%s' to be Synced and Ready\n", argocdNamespace, PostgresqlInstanceName)
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
	fmt.Printf("[%s] TEST PASSED - PostgreSQL Instance '%s' is Synced and Ready\n", argocdNamespace, PostgresqlInstanceName)
}

func testSecurityGroupSyncedAndReady(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Waiting for SecurityGroup related to '%s' to be Synced and Ready\n", argocdNamespace, PostgresqlInstanceName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for SecurityGroup related to '%s'", argocdNamespace, PostgresqlInstanceName), 60, 10*time.Second, func() (string, error) {
		sgName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SecurityGroupKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlInstanceName), "-o", "jsonpath={.items[0].metadata.name}")
		if err != nil {
			return "", err
		}
		if sgName == "" {
			return "", fmt.Errorf("no SecurityGroup found for composite '%s'", PostgresqlInstanceName)
		}
		expectedPrefix := PostgresqlInstanceName + "-sg-"
		if !strings.HasPrefix(sgName, expectedPrefix) {
			return "", fmt.Errorf("SecurityGroup name '%s' does not start with expected prefix '%s'", sgName, expectedPrefix)
		}
		syncStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SecurityGroupKind, sgName, "-o", `jsonpath={.status.conditions[?(@.type=="Synced")].status}`)
		if err != nil {
			return "", err
		}
		if syncStatus != "True" {
			return "", fmt.Errorf("SecurityGroup '%s' not synced yet, condition: %s", sgName, syncStatus)
		}
		readyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SecurityGroupKind, sgName, "-o", `jsonpath={.status.conditions[?(@.type=="Ready")].status}`)
		if err != nil {
			return "", err
		}
		if readyStatus != "True" {
			return "", fmt.Errorf("SecurityGroup '%s' not ready yet, condition: %s", sgName, readyStatus)
		}
		return sgName, nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] SecurityGroup for '%s' failed to become Synced and Ready", argocdNamespace, PostgresqlInstanceName))
	fmt.Printf("[%s] TEST PASSED - SecurityGroup is Synced and Ready\n", argocdNamespace)
}

func testSecurityGroupRulesSyncedAndReady(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Waiting for SecurityGroupRules related to '%s' to be Synced and Ready\n", argocdNamespace, PostgresqlInstanceName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for SecurityGroupRules related to '%s'", argocdNamespace, PostgresqlInstanceName), 60, 10*time.Second, func() (string, error) {
		ruleNames, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SecurityGroupRuleKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlInstanceName), "-o", "jsonpath={.items[*].metadata.name}")
		if err != nil {
			return "", err
		}
		names := strings.Fields(ruleNames)
		if len(names) < 2 {
			return "", fmt.Errorf("expected at least 2 SecurityGroupRules for composite '%s', found %d", PostgresqlInstanceName, len(names))
		}

		foundIngress := false
		foundEgress := false
		for _, name := range names {
			if strings.Contains(name, "-sg-ingress-") {
				foundIngress = true
			}
			if strings.Contains(name, "-sg-egress-") {
				foundEgress = true
			}
		}
		if !foundIngress {
			return "", fmt.Errorf("no ingress SecurityGroupRule found (expected name containing '-sg-ingress-')")
		}
		if !foundEgress {
			return "", fmt.Errorf("no egress SecurityGroupRule found (expected name containing '-sg-egress-')")
		}

		for _, name := range names {
			syncStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SecurityGroupRuleKind, name, "-o", `jsonpath={.status.conditions[?(@.type=="Synced")].status}`)
			if err != nil {
				return "", err
			}
			if syncStatus != "True" {
				return "", fmt.Errorf("SecurityGroupRule '%s' not synced yet, condition: %s", name, syncStatus)
			}
			readyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SecurityGroupRuleKind, name, "-o", `jsonpath={.status.conditions[?(@.type=="Ready")].status}`)
			if err != nil {
				return "", err
			}
			if readyStatus != "True" {
				return "", fmt.Errorf("SecurityGroupRule '%s' not ready yet, condition: %s", name, readyStatus)
			}
		}
		return "Synced+Ready", nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] SecurityGroupRules for '%s' failed to become Synced and Ready", argocdNamespace, PostgresqlInstanceName))
	fmt.Printf("[%s] TEST PASSED - SecurityGroupRules (ingress + egress) are Synced and Ready\n", argocdNamespace)
}

func testExternalSecretReady(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Waiting for ExternalSecret related to '%s' to be Ready\n", argocdNamespace, PostgresqlInstanceName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for ExternalSecret related to '%s'", argocdNamespace, PostgresqlInstanceName), 60, 10*time.Second, func() (string, error) {
		esName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", ExternalSecretKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlInstanceName), "-o", "jsonpath={.items[0].metadata.name}")
		if err != nil {
			return "", err
		}
		if esName == "" {
			return "", fmt.Errorf("no ExternalSecret found for composite '%s'", PostgresqlInstanceName)
		}
		expectedPrefix := PostgresqlInstanceName + "-es-"
		if !strings.HasPrefix(esName, expectedPrefix) {
			return "", fmt.Errorf("ExternalSecret name '%s' does not start with expected prefix '%s'", esName, expectedPrefix)
		}
		readyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", ExternalSecretKind, esName, "-o", `jsonpath={.status.conditions[?(@.type=="Ready")].status}`)
		if err != nil {
			return "", err
		}
		if readyStatus != "True" {
			return "", fmt.Errorf("ExternalSecret '%s' not ready yet, condition: %s", esName, readyStatus)
		}
		return esName, nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] ExternalSecret for '%s' failed to become Ready", argocdNamespace, PostgresqlInstanceName))

	// Verify ExternalSecret target template contains the username key (non-password key)
	esName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", ExternalSecretKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlInstanceName), "-o", "jsonpath={.items[0].metadata.name}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to find ExternalSecret", argocdNamespace))

	username, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", ExternalSecretKind, esName, "-o", "jsonpath={.spec.target.template.data.username}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get ExternalSecret target.template.data.username", argocdNamespace))
	require.Equal(t, "dbadmin", username, fmt.Sprintf("[%s] ExternalSecret '%s' username mismatch", argocdNamespace, esName))

	fmt.Printf("[%s] TEST PASSED - ExternalSecret '%s' is Ready, username=dbadmin\n", argocdNamespace, esName)
}

func testProviderConfigExists(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Verifying ProviderConfig '%s' exists\n", argocdNamespace, ProviderConfigExpectedName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for ProviderConfig '%s'", argocdNamespace, ProviderConfigExpectedName), 60, 10*time.Second, func() (string, error) {
		output, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlProviderConfigKind, ProviderConfigExpectedName, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if output == "" {
			return "", fmt.Errorf("ProviderConfig '%s' not found", ProviderConfigExpectedName)
		}
		return output, nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] ProviderConfig '%s' not found", argocdNamespace, ProviderConfigExpectedName))
	fmt.Printf("[%s] TEST PASSED - ProviderConfig '%s' exists\n", argocdNamespace, ProviderConfigExpectedName)
}

func testRdsInstanceSyncedAndReady(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Waiting for RDS Instance related to '%s' to be Synced and Ready\n", argocdNamespace, PostgresqlInstanceName)
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
	fmt.Printf("[%s] TEST PASSED - RDS Instance is Synced and Ready\n", argocdNamespace)
}

func testRdsInstanceFieldsVerified(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	fmt.Printf("[%s] TEST: Verifying RDS Instance fields for '%s'\n", argocdNamespace, PostgresqlInstanceName)
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

	// Verify deletionProtection is false
	deletionProtection, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", RdsInstanceKind, rdsName, "-o", "jsonpath={.spec.forProvider.deletionProtection}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get deletionProtection", argocdNamespace))
	require.Equal(t, "false", deletionProtection, fmt.Sprintf("[%s] RDS Instance '%s' deletionProtection should be false", argocdNamespace, rdsName))

	// Verify Status fields propagation on PostgreSQLInstance (endpoint address and port)
	endpointAddress, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", PostgresqlInstanceKind, PostgresqlInstanceName, "-o", "jsonpath={.status.endpoint.address}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get endpoint address from PostgreSQLInstance status", argocdNamespace))
	require.NotEmpty(t, endpointAddress, fmt.Sprintf("[%s] PostgreSQLInstance '%s' status endpoint address is empty", argocdNamespace, PostgresqlInstanceName))

	endpointPort, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", PostgresqlInstanceKind, PostgresqlInstanceName, "-o", "jsonpath={.status.endpoint.port}")
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to get endpoint port from PostgreSQLInstance status", argocdNamespace))
	require.NotEmpty(t, endpointPort, fmt.Sprintf("[%s] PostgreSQLInstance '%s' status endpoint port is empty", argocdNamespace, PostgresqlInstanceName))

	fmt.Printf("[%s] TEST PASSED - RDS Instance fields verified (allocatedStorage=20, engineVersion=17.2, instanceClass=db.t3.micro, deletionProtection=false), status propagation OK (endpoint=%s, port=%s)\n", argocdNamespace, endpointAddress, endpointPort)
}

func testDeletionProtectionToggle(t *testing.T, argocdNamespace string, namespaceOptions *terrak8s.KubectlOptions) {
	// Step 1: Patch PostgreSQLInstance deletionProtection to true
	fmt.Printf("[%s] TEST: Patching PostgreSQL Instance '%s' deletionProtection to true\n", argocdNamespace, PostgresqlInstanceName)
	_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "patch", PostgresqlInstanceKind, PostgresqlInstanceName, "-n", PostgresqlNamespaceName, "--type", "merge", "-p", `{"spec":{"deletionProtection":true}}`)
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to patch deletionProtection to true", argocdNamespace))

	// Step 2: Wait for RDS Instance deletionProtection to become true
	fmt.Printf("[%s] TEST: Waiting for RDS Instance deletionProtection to propagate to true\n", argocdNamespace)
	_, err = retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for RDS deletionProtection=true", argocdNamespace), 30, 10*time.Second, func() (string, error) {
		rdsName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", RdsInstanceKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlInstanceName), "-o", "jsonpath={.items[0].metadata.name}")
		if err != nil {
			return "", err
		}
		dp, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", RdsInstanceKind, rdsName, "-o", "jsonpath={.spec.forProvider.deletionProtection}")
		if err != nil {
			return "", err
		}
		if dp != "true" {
			return "", fmt.Errorf("RDS deletionProtection is '%s', expected 'true'", dp)
		}
		return "true", nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] RDS Instance deletionProtection failed to propagate to true", argocdNamespace))
	fmt.Printf("[%s] TEST PASSED - RDS Instance deletionProtection propagated to true\n", argocdNamespace)

	// Step 3: Patch PostgreSQLInstance deletionProtection back to false
	fmt.Printf("[%s] TEST: Patching PostgreSQL Instance '%s' deletionProtection back to false\n", argocdNamespace, PostgresqlInstanceName)
	_, err = terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "patch", PostgresqlInstanceKind, PostgresqlInstanceName, "-n", PostgresqlNamespaceName, "--type", "merge", "-p", `{"spec":{"deletionProtection":false}}`)
	require.NoError(t, err, fmt.Sprintf("[%s] Failed to patch deletionProtection to false", argocdNamespace))

	// Wait for RDS Instance deletionProtection to propagate back to false
	_, err = retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for RDS deletionProtection=false", argocdNamespace), 30, 10*time.Second, func() (string, error) {
		currentRdsName, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", RdsInstanceKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlInstanceName), "-o", "jsonpath={.items[0].metadata.name}")
		if err != nil {
			return "", err
		}
		dp, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", RdsInstanceKind, currentRdsName, "-o", "jsonpath={.spec.forProvider.deletionProtection}")
		if err != nil {
			return "", err
		}
		if dp != "false" {
			return "", fmt.Errorf("RDS deletionProtection is '%s', expected 'false'", dp)
		}
		return "false", nil
	})
	require.NoError(t, err, fmt.Sprintf("[%s] RDS Instance deletionProtection failed to propagate back to false", argocdNamespace))

	fmt.Printf("[%s] TEST PASSED - Deletion protection toggle verified (false -> true -> false)\n", argocdNamespace)
}
