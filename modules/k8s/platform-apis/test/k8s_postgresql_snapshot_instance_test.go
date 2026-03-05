package test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsrds "github.com/aws/aws-sdk-go/service/rds"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
)

const (
	PostgresqlSnapshotInstanceName = "postgresql-instance-from-snapshot-test"
)

func runPostgresqlSnapshotInstanceTests(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	snapshotID := testFindAutomatedSnapshot(t, namespaceOptions)

	testPostgresqlSnapshotInstanceApplied(t, namespaceOptions, snapshotID)
	t.Run("snapshot-sub-resources", func(t *testing.T) {
		t.Run("security-group-rules", func(t *testing.T) {
			t.Parallel()
			testSnapshotSecurityGroupRulesSyncedAndReady(t, namespaceOptions)
		})
		t.Run("security-group", func(t *testing.T) {
			t.Parallel()
			testSnapshotSecurityGroupSyncedAndReady(t, namespaceOptions)
		})
		t.Run("external-secret", func(t *testing.T) {
			t.Parallel()
			testSnapshotExternalSecretReady(t, namespaceOptions)
		})
		t.Run("provider-config", func(t *testing.T) {
			t.Parallel()
			testSnapshotProviderConfigExists(t, namespaceOptions)
		})
		t.Run("rds-snapshot-instance", func(t *testing.T) {
			t.Parallel()
			waitSyncedAndReadyByLabel(t, namespaceOptions, RdsInstanceKind, PostgresqlSnapshotInstanceName, 90, 10*time.Second)
		})
	})

	if t.Failed() {
		return
	}
	testPostgresqlSnapshotInstanceSyncedAndReady(t, namespaceOptions)
	testSnapshotInstanceVerifiedFromSnapshot(t, namespaceOptions, snapshotID)
	testSnapshotIdentifierImmutable(t, namespaceOptions)
}

func testFindAutomatedSnapshot(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) string {
	rdsName, err := getFirstByLabel(t, namespaceOptions, RdsInstanceKind, PostgresqlInstanceName)
	require.NoError(t, err, "failed to find RDS instance for %s", PostgresqlInstanceName)
	require.NotEmpty(t, rdsName, "no RDS instance found for %s", PostgresqlInstanceName)

	dbIdentifier, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", RdsInstanceKind, rdsName, "-o", "jsonpath={.spec.forProvider.identifier}")
	require.NoError(t, err, "failed to get RDS instance identifier")
	require.NotEmpty(t, dbIdentifier, "RDS instance identifier is empty")

	region, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", RdsInstanceKind, rdsName, "-o", "jsonpath={.spec.forProvider.region}")
	require.NoError(t, err, "failed to get RDS instance region")
	require.NotEmpty(t, region, "RDS instance region is empty")

	var snapshotID string
	_, err = retry.DoWithRetryE(t, fmt.Sprintf("waiting for automated snapshot for %s", dbIdentifier), 60, 2*time.Minute, func() (string, error) {
		id, findErr := findLatestAutomatedSnapshot(dbIdentifier, region)
		if findErr != nil {
			return "", findErr
		}
		snapshotID = id
		return id, nil
	})
	require.NoError(t, err, "no automated snapshot found for RDS instance %s", dbIdentifier)
	return snapshotID
}

func findLatestAutomatedSnapshot(dbIdentifier, region string) (string, error) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return "", fmt.Errorf("failed to create AWS session: %w", err)
	}
	svc := awsrds.New(sess)
	resp, err := svc.DescribeDBSnapshots(&awsrds.DescribeDBSnapshotsInput{
		DBInstanceIdentifier: aws.String(dbIdentifier),
		SnapshotType:         aws.String("automated"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to describe RDS snapshots for %s: %w", dbIdentifier, err)
	}
	var latest *awsrds.DBSnapshot
	for _, s := range resp.DBSnapshots {
		if s.Status == nil || *s.Status != "available" || s.DBSnapshotIdentifier == nil {
			continue
		}
		if latest == nil || (s.SnapshotCreateTime != nil && latest.SnapshotCreateTime != nil && s.SnapshotCreateTime.After(*latest.SnapshotCreateTime)) {
			latest = s
		}
	}
	if latest == nil {
		return "", fmt.Errorf("no available automated snapshots for DB instance %s", dbIdentifier)
	}
	return *latest.DBSnapshotIdentifier, nil
}

func testPostgresqlSnapshotInstanceApplied(t *testing.T, namespaceOptions *terrak8s.KubectlOptions, snapshotID string) {
	yaml := fmt.Sprintf(`apiVersion: database.entigo.com/v1alpha1
kind: PostgreSQLInstance
metadata:
  name: %s
  namespace: %s
spec:
  allocatedStorage: 20
  engineVersion: "17.2"
  instanceType: db.t3.micro
  allowMajorVersionUpgrade: false
  autoMinorVersionUpgrade: true
  deletionProtection: false
  multiAZ: false
  snapshotIdentifier: %s
`, PostgresqlSnapshotInstanceName, PostgresqlNamespaceName, snapshotID)

	tmpFile, err := os.CreateTemp("", "snapshot-instance-*.yaml")
	require.NoError(t, err, "failed to create temp file for snapshot instance")
	defer os.Remove(tmpFile.Name())

	_, err = fmt.Fprint(tmpFile, yaml)
	require.NoError(t, err, "failed to write snapshot instance YAML")
	require.NoError(t, tmpFile.Close())

	_, err = terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "apply", "-f", tmpFile.Name())
	require.NoError(t, err, "applying PostgreSQL snapshot instance error")
}

func testPostgresqlSnapshotInstanceSyncedAndReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	waitSyncedAndReady(t, namespaceOptions, PostgresqlInstanceKind, PostgresqlSnapshotInstanceName, 90, 10*time.Second)
}

func testSnapshotSecurityGroupRulesSyncedAndReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for SecurityGroupRules (composite=%s)", PostgresqlSnapshotInstanceName), 60, 10*time.Second, func() (string, error) {
		ruleNames, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SecurityGroupRuleKind, "-l",
			fmt.Sprintf("crossplane.io/composite=%s", PostgresqlSnapshotInstanceName), "-o", "jsonpath={.items[*].metadata.name}")
		if err != nil {
			return "", err
		}
		names := strings.Fields(ruleNames)
		if len(names) < 2 {
			return "", fmt.Errorf("expected at least 2 SecurityGroupRules for composite=%s, found %d", PostgresqlSnapshotInstanceName, len(names))
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
	require.NoError(t, err, fmt.Sprintf("SecurityGroupRules for '%s' failed to become Synced and Ready", PostgresqlSnapshotInstanceName))
}

func testSnapshotSecurityGroupSyncedAndReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for SecurityGroup (composite=%s)", PostgresqlSnapshotInstanceName), 60, 10*time.Second, func() (string, error) {
		sgName, err := getFirstByLabel(t, namespaceOptions, SecurityGroupKind, PostgresqlSnapshotInstanceName)
		if err != nil {
			return "", err
		}
		if sgName == "" {
			return "", fmt.Errorf("no SecurityGroup found for composite=%s", PostgresqlSnapshotInstanceName)
		}
		expectedPrefix := PostgresqlSnapshotInstanceName + "-sg-"
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
	require.NoError(t, err, fmt.Sprintf("SecurityGroup for '%s' failed to become Synced and Ready", PostgresqlSnapshotInstanceName))
}

func testSnapshotExternalSecretReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for ExternalSecret (composite=%s)", PostgresqlSnapshotInstanceName), 90, 10*time.Second, func() (string, error) {
		esName, err := getFirstByLabel(t, namespaceOptions, ExternalSecretKind, PostgresqlSnapshotInstanceName)
		if err != nil {
			return "", err
		}
		if esName == "" {
			return "", fmt.Errorf("no ExternalSecret found for composite=%s", PostgresqlSnapshotInstanceName)
		}
		expectedPrefix := PostgresqlSnapshotInstanceName + "-es-"
		if !strings.HasPrefix(esName, expectedPrefix) {
			return "", fmt.Errorf("ExternalSecret name '%s' does not start with expected prefix '%s'", esName, expectedPrefix)
		}
		readyStatus, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", ExternalSecretKind, esName, "-o",
			`jsonpath={.status.conditions[?(@.type=="Ready")].status}`)
		if err != nil {
			return "", err
		}
		if readyStatus != "True" {
			return "", fmt.Errorf("ExternalSecret '%s' not ready yet, condition: %s", esName, readyStatus)
		}
		return esName, nil
	})
	require.NoError(t, err, fmt.Sprintf("ExternalSecret for '%s' failed to become Ready", PostgresqlSnapshotInstanceName))
}

func testSnapshotProviderConfigExists(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	expectedName := PostgresqlSnapshotInstanceName + "-providerconfig"
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for ProviderConfig '%s'", expectedName), 90, 10*time.Second, func() (string, error) {
		output, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", SqlProviderConfigKind, expectedName, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if output == "" {
			return "", fmt.Errorf("ProviderConfig '%s' not found", expectedName)
		}
		return output, nil
	})
	require.NoError(t, err, fmt.Sprintf("ProviderConfig '%s' not found", expectedName))
}

func testSnapshotInstanceVerifiedFromSnapshot(t *testing.T, namespaceOptions *terrak8s.KubectlOptions, expectedSnapshotID string) {
	rdsName, err := getFirstByLabel(t, namespaceOptions, RdsInstanceKind, PostgresqlSnapshotInstanceName)
	require.NoError(t, err, "failed to find RDS instance for %s", PostgresqlSnapshotInstanceName)
	require.NotEmpty(t, rdsName, "no RDS instance found for %s", PostgresqlSnapshotInstanceName)

	snapshotID, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", RdsInstanceKind, rdsName, "-o", "jsonpath={.spec.forProvider.snapshotIdentifier}")
	require.NoError(t, err, "failed to get snapshotIdentifier from RDS instance")
	require.Equal(t, expectedSnapshotID, snapshotID, "RDS instance snapshotIdentifier does not match the expected snapshot")
}

func testSnapshotIdentifierImmutable(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	output, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "patch", PostgresqlInstanceKind, PostgresqlSnapshotInstanceName,
		"-n", PostgresqlNamespaceName, "--type", "merge", "-p", `{"spec":{"snapshotIdentifier":"rds:different-snapshot-id"}}`)
	require.Error(t, err, "patching snapshotIdentifier to a different value should be rejected")
	require.Contains(t, output, "snapshotIdentifier is immutable", "rejection message should mention immutability")
}
