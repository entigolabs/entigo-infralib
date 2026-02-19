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

func runPostgresqlInstanceTests(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	testPostgresqlInstanceApplied(t, namespaceOptions)
	t.Run("sub-resources", func(t *testing.T) {
		t.Run("security-group-rules", func(t *testing.T) {
			t.Parallel()
			testSecurityGroupRulesSyncedAndReady(t, namespaceOptions)
		})
		t.Run("security-group", func(t *testing.T) {
			t.Parallel()
			testSecurityGroupSyncedAndReady(t, namespaceOptions)
		})
		t.Run("external-secret", func(t *testing.T) {
			t.Parallel()
			testExternalSecretReady(t, namespaceOptions)
		})
		t.Run("provider-config", func(t *testing.T) {
			t.Parallel()
			testProviderConfigExists(t, namespaceOptions)
		})
		t.Run("rds-instance", func(t *testing.T) {
			t.Parallel()
			testRdsInstanceSyncedAndReady(t, namespaceOptions)
		})
	})

	if t.Failed() {
		return
	}
	testPostgresqlInstanceSyncedAndReady(t, namespaceOptions)
	testRdsInstanceFieldsVerified(t, namespaceOptions)
	testDeletionProtectionToggle(t, namespaceOptions)
}

func testPostgresqlInstanceApplied(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "apply", "-f", "./templates/postgresql_test_instance.yaml", "-n", PostgresqlNamespaceName)
	require.NoError(t, err, "applying PostgreSQL Instance error")
}

func testPostgresqlInstanceSyncedAndReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	waitSyncedAndReady(t, namespaceOptions, PostgresqlInstanceKind, PostgresqlInstanceName, 90, 10*time.Second)
}

func testSecurityGroupSyncedAndReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for SecurityGroup (composite=%s)", PostgresqlInstanceName), 60, 10*time.Second, func() (string, error) {
		sgName, err := getFirstByLabel(t, namespaceOptions, SecurityGroupKind, PostgresqlInstanceName)
		if err != nil {
			return "", err
		}
		if sgName == "" {
			return "", fmt.Errorf("no SecurityGroup found for composite=%s", PostgresqlInstanceName)
		}
		expectedPrefix := PostgresqlInstanceName + "-sg-"
		if !strings.HasPrefix(sgName, expectedPrefix) {
			return "", fmt.Errorf("SecurityGroup name '%s' does not start with expected prefix '%s'", sgName, expectedPrefix)
		}
		for _, condType := range []string{"Synced", "Ready"} {
			status, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SecurityGroupKind, sgName, "-o",
				fmt.Sprintf(`jsonpath={.status.conditions[?(@.type=="%s")].status}`, condType))
			if err != nil {
				return "", err
			}
			if status != "True" {
				return "", fmt.Errorf("SecurityGroup '%s': %s=%s", sgName, condType, status)
			}
		}
		return sgName, nil
	})
	require.NoError(t, err, fmt.Sprintf("SecurityGroup for '%s' failed to become Synced and Ready", PostgresqlInstanceName))
}

func testSecurityGroupRulesSyncedAndReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for SecurityGroupRules (composite=%s)", PostgresqlInstanceName), 60, 10*time.Second, func() (string, error) {
		ruleNames, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SecurityGroupRuleKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlInstanceName), "-o", "jsonpath={.items[*].metadata.name}")
		if err != nil {
			return "", err
		}
		names := strings.Fields(ruleNames)
		if len(names) < 2 {
			return "", fmt.Errorf("expected at least 2 SecurityGroupRules for composite=%s, found %d", PostgresqlInstanceName, len(names))
		}

		foundIngress, foundEgress := false, false
		for _, name := range names {
			if strings.Contains(name, "-sg-ingress-") {
				foundIngress = true
			}
			if strings.Contains(name, "-sg-egress-") {
				foundEgress = true
			}
		}
		if !foundIngress {
			return "", fmt.Errorf("no ingress SecurityGroupRule found")
		}
		if !foundEgress {
			return "", fmt.Errorf("no egress SecurityGroupRule found")
		}

		for _, name := range names {
			for _, condType := range []string{"Synced", "Ready"} {
				status, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SecurityGroupRuleKind, name, "-o",
					fmt.Sprintf(`jsonpath={.status.conditions[?(@.type=="%s")].status}`, condType))
				if err != nil {
					return "", err
				}
				if status != "True" {
					return "", fmt.Errorf("SecurityGroupRule '%s': %s=%s", name, condType, status)
				}
			}
		}
		return "Synced+Ready", nil
	})
	require.NoError(t, err, fmt.Sprintf("SecurityGroupRules for '%s' failed to become Synced and Ready", PostgresqlInstanceName))
}

func testExternalSecretReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for ExternalSecret (composite=%s)", PostgresqlInstanceName), 60, 10*time.Second, func() (string, error) {
		esName, err := getFirstByLabel(t, namespaceOptions, ExternalSecretKind, PostgresqlInstanceName)
		if err != nil {
			return "", err
		}
		if esName == "" {
			return "", fmt.Errorf("no ExternalSecret found for composite=%s", PostgresqlInstanceName)
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
	require.NoError(t, err, fmt.Sprintf("ExternalSecret for '%s' failed to become Ready", PostgresqlInstanceName))

	esName, err := getFirstByLabel(t, namespaceOptions, ExternalSecretKind, PostgresqlInstanceName)
	require.NoError(t, err, "failed to find ExternalSecret")

	username, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", ExternalSecretKind, esName, "-o", "jsonpath={.spec.target.template.data.username}")
	require.NoError(t, err, "failed to get ExternalSecret username field")
	require.Equal(t, "dbadmin", username, fmt.Sprintf("ExternalSecret '%s' username mismatch", esName))
}

func testProviderConfigExists(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for ProviderConfig '%s'", ProviderConfigExpectedName), 60, 10*time.Second, func() (string, error) {
		output, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlProviderConfigKind, ProviderConfigExpectedName, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if output == "" {
			return "", fmt.Errorf("ProviderConfig '%s' not found", ProviderConfigExpectedName)
		}
		return output, nil
	})
	require.NoError(t, err, fmt.Sprintf("ProviderConfig '%s' not found", ProviderConfigExpectedName))
}

func testRdsInstanceSyncedAndReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	waitSyncedAndReadyByLabel(t, namespaceOptions, RdsInstanceKind, PostgresqlInstanceName, 60, 10*time.Second)
}

func testRdsInstanceFieldsVerified(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	rdsName, err := getFirstByLabel(t, namespaceOptions, RdsInstanceKind, PostgresqlInstanceName)
	require.NoError(t, err, "failed to find RDS Instance")
	require.NotEmpty(t, rdsName, fmt.Sprintf("no RDS Instance found for composite '%s'", PostgresqlInstanceName))

	allocatedStorage, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", RdsInstanceKind, rdsName, "-o", "jsonpath={.spec.forProvider.allocatedStorage}")
	require.NoError(t, err, "failed to get allocatedStorage")
	require.Equal(t, "20", allocatedStorage, fmt.Sprintf("RDS Instance '%s' allocatedStorage mismatch", rdsName))

	engineVersion, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", RdsInstanceKind, rdsName, "-o", "jsonpath={.spec.forProvider.engineVersion}")
	require.NoError(t, err, "failed to get engineVersion")
	require.Equal(t, "17.2", engineVersion, fmt.Sprintf("RDS Instance '%s' engineVersion mismatch", rdsName))

	instanceClass, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", RdsInstanceKind, rdsName, "-o", "jsonpath={.spec.forProvider.instanceClass}")
	require.NoError(t, err, "failed to get instanceClass")
	require.Equal(t, "db.t3.micro", instanceClass, fmt.Sprintf("RDS Instance '%s' instanceClass mismatch", rdsName))

	deletionProtection, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", RdsInstanceKind, rdsName, "-o", "jsonpath={.spec.forProvider.deletionProtection}")
	require.NoError(t, err, "failed to get deletionProtection")
	require.Equal(t, "false", deletionProtection, fmt.Sprintf("RDS Instance '%s' deletionProtection should be false", rdsName))

	endpointAddress, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", PostgresqlInstanceKind, PostgresqlInstanceName, "-o", "jsonpath={.status.endpoint.address}")
	require.NoError(t, err, "failed to get endpoint address from PostgreSQLInstance status")
	require.NotEmpty(t, endpointAddress, fmt.Sprintf("PostgreSQLInstance '%s' status endpoint address is empty", PostgresqlInstanceName))

	endpointPort, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", PostgresqlInstanceKind, PostgresqlInstanceName, "-o", "jsonpath={.status.endpoint.port}")
	require.NoError(t, err, "failed to get endpoint port from PostgreSQLInstance status")
	require.NotEmpty(t, endpointPort, fmt.Sprintf("PostgreSQLInstance '%s' status endpoint port is empty", PostgresqlInstanceName))
}

func testDeletionProtectionToggle(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "patch", PostgresqlInstanceKind, PostgresqlInstanceName, "-n", PostgresqlNamespaceName, "--type", "merge", "-p", `{"spec":{"deletionProtection":true}}`)
	require.NoError(t, err, "failed to patch deletionProtection to true")

	_, err = retry.DoWithRetryE(t, "waiting for RDS deletionProtection=true", 30, 10*time.Second, func() (string, error) {
		rdsName, err := getFirstByLabel(t, namespaceOptions, RdsInstanceKind, PostgresqlInstanceName)
		if err != nil {
			return "", err
		}
		if rdsName == "" {
			return "", fmt.Errorf("no RDS Instance found for composite=%s", PostgresqlInstanceName)
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
	require.NoError(t, err, "RDS Instance deletionProtection failed to propagate to true")

	_, err = terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "patch", PostgresqlInstanceKind, PostgresqlInstanceName, "-n", PostgresqlNamespaceName, "--type", "merge", "-p", `{"spec":{"deletionProtection":false}}`)
	require.NoError(t, err, "failed to patch deletionProtection to false")

	_, err = retry.DoWithRetryE(t, "waiting for RDS deletionProtection=false", 30, 10*time.Second, func() (string, error) {
		rdsName, err := getFirstByLabel(t, namespaceOptions, RdsInstanceKind, PostgresqlInstanceName)
		if err != nil {
			return "", err
		}
		if rdsName == "" {
			return "", fmt.Errorf("no RDS Instance found for composite=%s", PostgresqlInstanceName)
		}
		dp, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", RdsInstanceKind, rdsName, "-o", "jsonpath={.spec.forProvider.deletionProtection}")
		if err != nil {
			return "", err
		}
		if dp != "false" {
			return "", fmt.Errorf("RDS deletionProtection is '%s', expected 'false'", dp)
		}
		return "false", nil
	})
	require.NoError(t, err, "RDS Instance deletionProtection failed to propagate back to false")
}
