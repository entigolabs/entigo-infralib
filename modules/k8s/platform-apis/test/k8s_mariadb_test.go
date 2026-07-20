package test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
)

func testMariadb(t *testing.T, ctx context.Context, cluster, argocd *terrak8s.KubectlOptions) {
	mdbNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, MariadbNamespaceName)
	defer cleanupMariadb(t, cluster, argocd)

	if ctx.Err() != nil {
		return
	}
	applyFile(t, cluster, "./templates/mariadb_test_application.yaml")
	syncWithRetry(t, argocd, MariadbApplicationName)
	if ctx.Err() != nil {
		return
	}

	t.Run("Instance", func(t *testing.T) { testMariadbInstance(t, mdbNs) })
	if t.Failed() {
		return
	}

	t.Run("Lifecycle", func(t *testing.T) { testMariadbLifecycle(t, mdbNs) })
}

func testMariadbInstance(t *testing.T, mdbNs *terrak8s.KubectlOptions) {
	t.Helper()

	t.Run("SubResources", func(t *testing.T) {
		t.Run("RdsInstance", func(t *testing.T) {
			t.Parallel()
			waitSyncedAndReadyByLabel(t, mdbNs, RdsInstanceKind, MariadbInstanceName, 60, 10*time.Second)
		})
		t.Run("SecurityGroup", func(t *testing.T) {
			t.Parallel()
			waitSyncedAndReadyByLabel(t, mdbNs, SecurityGroupKind, MariadbInstanceName, 60, 10*time.Second)
		})
		t.Run("SecurityGroupRules", func(t *testing.T) {
			t.Parallel()
			waitMariadbSecurityGroupRulesReady(t, mdbNs)
		})
		t.Run("ExternalSecret", func(t *testing.T) {
			t.Parallel()
			waitMariadbExternalSecretReady(t, mdbNs)
		})
		t.Run("ProviderConfig", func(t *testing.T) {
			t.Parallel()
			waitResourceExists(t, mdbNs, MariadbSqlProviderConfigKind, MariadbInstanceName+"-providerconfig", 90, 10*time.Second)
		})
	})
	if t.Failed() {
		return
	}

	waitSyncedAndReady(t, mdbNs, MariadbInstanceKind, MariadbInstanceName, 90, 10*time.Second)
	if t.Failed() {
		return
	}

	rdsName, err := getFirstByLabel(t, mdbNs, RdsInstanceKind, MariadbInstanceName)
	require.NoError(t, err)
	require.NotEmpty(t, rdsName)

	// RDS fields must reflect what was specified on the composite
	require.Equal(t, "20", getField(t, mdbNs, RdsInstanceKind, rdsName, ".status.atProvider.allocatedStorage"))
	require.Equal(t, "11.4.10", getField(t, mdbNs, RdsInstanceKind, rdsName, ".status.atProvider.engineVersion"))
	require.Equal(t, "db.t3.micro", getField(t, mdbNs, RdsInstanceKind, rdsName, ".status.atProvider.instanceClass"))
	require.Equal(t, "false", getField(t, mdbNs, RdsInstanceKind, rdsName, ".status.atProvider.deletionProtection"))

	// Composite status endpoint must be populated once instance is ready
	require.NotEmpty(t, getField(t, mdbNs, MariadbInstanceKind, MariadbInstanceName, ".status.endpoint.address"),
		"endpoint address should be populated")
	require.NotEmpty(t, getField(t, mdbNs, MariadbInstanceKind, MariadbInstanceName, ".status.endpoint.port"),
		"endpoint port should be populated")

	// ExternalSecret must carry the admin username
	esName, err := getFirstByLabel(t, mdbNs, ExternalSecretKind, MariadbInstanceName)
	require.NoError(t, err)
	require.Equal(t, "dbadmin", getField(t, mdbNs, ExternalSecretKind, esName, ".spec.target.template.data.username"))

	// deletionProtection on composite propagates to RDS spec
	patchResource(t, mdbNs, MariadbInstanceKind, MariadbInstanceName, `{"spec":{"deletionProtection":true}}`)
	waitFieldEquals(t, mdbNs, RdsInstanceKind, rdsName, ".spec.forProvider.deletionProtection", "true", 30, 10*time.Second)
	patchResource(t, mdbNs, MariadbInstanceKind, MariadbInstanceName, `{"spec":{"deletionProtection":false}}`)
	waitFieldEquals(t, mdbNs, RdsInstanceKind, rdsName, ".spec.forProvider.deletionProtection", "false", 30, 10*time.Second)
}

// testMariadbLifecycle drives a single MariaDBInstance through every engineVersion/
// parameterGroupParameters combination worth covering, since each transition is a real (slow) AWS
// RDS change and provisioning a separate instance per case would multiply e2e cost for no extra
// coverage.
func testMariadbLifecycle(t *testing.T, mdbNs *terrak8s.KubectlOptions) {
	t.Helper()

	waitSyncedAndReady(t, mdbNs, MariadbInstanceKind, MariadbLifecycleName, 120, 10*time.Second)
	if t.Failed() {
		return
	}

	rdsName, err := getFirstByLabel(t, mdbNs, RdsInstanceKind, MariadbLifecycleName)
	require.NoError(t, err)
	require.NotEmpty(t, rdsName)

	require.Equal(t, "10.11.16", getField(t, mdbNs, RdsInstanceKind, rdsName, ".spec.forProvider.engineVersion"))
	actualFamily := "mariadb10.11"

	_, err = getFirstByLabel(t, mdbNs, RdsParameterGroupKind, MariadbLifecycleName)
	require.Error(t, err, "no ParameterGroup should exist while parameterGroupParameters is unset")

	patchResource(t, mdbNs, MariadbInstanceKind, MariadbLifecycleName, `{"spec":{"parameterGroupParameters":{"applyMethod":"pending-reboot","max_connections":"200"}}}`)

	pgName := waitSyncedAndReadyByLabel(t, mdbNs, RdsParameterGroupKind, MariadbLifecycleName, 60, 10*time.Second)
	require.NotEmpty(t, pgName)
	require.Equal(t, actualFamily, getField(t, mdbNs, RdsParameterGroupKind, pgName, ".spec.forProvider.family"))
	require.Equal(t, "max_connections", getField(t, mdbNs, RdsParameterGroupKind, pgName, ".spec.forProvider.parameter[0].name"))
	require.Equal(t, "200", getField(t, mdbNs, RdsParameterGroupKind, pgName, ".spec.forProvider.parameter[0].value"))
	waitFieldEquals(t, mdbNs, RdsInstanceKind, rdsName, ".spec.forProvider.parameterGroupName", pgName, 60, 10*time.Second)

	patchResource(t, mdbNs, MariadbInstanceKind, MariadbLifecycleName, `{"spec":{"parameterGroupParameters":{"max_connections":"300"}}}`)

	waitFieldEquals(t, mdbNs, RdsParameterGroupKind, pgName, ".spec.forProvider.parameter[0].value", "300", 60, 10*time.Second)
	require.Equal(t, pgName, getField(t, mdbNs, RdsInstanceKind, rdsName, ".spec.forProvider.parameterGroupName"),
		"ParameterGroup should update in place, not be recreated under a different name")

	newVersion, newFamily := "11.4.10", "mariadb11.4"
	patchResource(t, mdbNs, MariadbInstanceKind, MariadbLifecycleName, `{"spec":{"engineVersion":"`+newVersion+`"}}`)

	recreatedPgName := waitSyncedAndReadyByLabelWhere(t, mdbNs, RdsParameterGroupKind, MariadbLifecycleName, ".spec.forProvider.family", newFamily, 60, 10*time.Second)
	require.NotEqual(t, pgName, recreatedPgName)

	waitFieldEquals(t, mdbNs, RdsInstanceKind, rdsName, ".status.atProvider.parameterGroupName", recreatedPgName, 240, 15*time.Second)
	waitFieldEquals(t, mdbNs, RdsInstanceKind, rdsName, ".spec.forProvider.parameterGroupName", recreatedPgName, 60, 10*time.Second)
	waitFieldEquals(t, mdbNs, RdsInstanceKind, rdsName, ".spec.forProvider.engineVersion", newVersion, 60, 10*time.Second)
	newOptionSuffix := strings.ReplaceAll(strings.TrimPrefix(newFamily, "mariadb"), ".", "-")
	waitFieldEquals(t, mdbNs, RdsInstanceKind, rdsName, ".spec.forProvider.optionGroupName", "default:mariadb-"+newOptionSuffix, 60, 10*time.Second)

	cleanupWaitGone(t, mdbNs, RdsParameterGroupKind, pgName, 30)

	patchResource(t, mdbNs, MariadbInstanceKind, MariadbLifecycleName, `{"spec":{"parameterGroupParameters":null}}`)

	cleanupWaitGone(t, mdbNs, RdsParameterGroupKind, recreatedPgName, 30)
	waitFieldEquals(t, mdbNs, RdsInstanceKind, rdsName, ".spec.forProvider.parameterGroupName", "default."+newFamily, 60, 10*time.Second)
}

func waitMariadbSecurityGroupRulesReady(t *testing.T, mdbNs *terrak8s.KubectlOptions) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("SecurityGroupRules for %s", MariadbInstanceName), 60, 10*time.Second,
		func() (string, error) {
			rules, err := getMariadbSecurityGroupRules(t, mdbNs)
			if err != nil {
				return "", err
			}
			if err := validateIngressEgressExists(rules); err != nil {
				return "", err
			}
			return checkAllRulesReady(t, mdbNs, rules)
		})
	require.NoError(t, err, "SecurityGroupRules for %s never became ready", MariadbInstanceName)
}

func getMariadbSecurityGroupRules(t *testing.T, mdbNs *terrak8s.KubectlOptions) ([]string, error) {
	names, err := terrak8s.RunKubectlAndGetOutputE(t, mdbNs, "get", SecurityGroupRuleKind,
		"-l", fmt.Sprintf("crossplane.io/composite=%s", MariadbInstanceName),
		"-o", "jsonpath={.items[*].metadata.name}")
	if err != nil {
		return nil, err
	}
	rules := strings.Fields(names)
	if len(rules) < 2 {
		return nil, fmt.Errorf("expected ≥2 rules, got %d", len(rules))
	}
	return rules, nil
}

func waitMariadbExternalSecretReady(t *testing.T, mdbNs *terrak8s.KubectlOptions) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("ExternalSecret for %s", MariadbInstanceName), 90, 10*time.Second,
		func() (string, error) {
			name, err := getFirstByLabel(t, mdbNs, ExternalSecretKind, MariadbInstanceName)
			if err != nil || name == "" {
				return "", fmt.Errorf("ExternalSecret not found yet")
			}
			if !strings.HasPrefix(name, MariadbInstanceName+"-es-") {
				return "", fmt.Errorf("unexpected ExternalSecret name: %s", name)
			}
			return checkConditions(t, mdbNs, ExternalSecretKind, name, "Ready")
		})
	require.NoError(t, err)
}

func cleanupMariadb(t *testing.T, cluster, argocd *terrak8s.KubectlOptions) {
	if t.Failed() {
		return // leave resources in place for debugging
	}
	mdbNs := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, MariadbNamespaceName)

	cleanupDisableDeletionProtectionOnMariadbInstance(t, mdbNs)
	cleanupDeleteParallel(t, mdbNs, MariadbInstanceKind, 180, MariadbInstanceName, MariadbLifecycleName)

	_, _ = terrak8s.RunKubectlAndGetOutputE(t, argocd, "delete", "application", MariadbApplicationName, "--ignore-not-found")
}

func cleanupDisableDeletionProtectionOnMariadbInstance(t *testing.T, mdbNs *terrak8s.KubectlOptions) {
	patchDeletionProtectionIfEnabled(t, mdbNs, MariadbInstanceKind, MariadbInstanceName)

	_, _ = retry.DoWithRetryE(t, "waiting for RDS deletionProtection=false", 30, 10*time.Second,
		func() (string, error) {
			rdsName, err := getFirstByLabel(t, mdbNs, RdsInstanceKind, MariadbInstanceName)
			if err != nil || rdsName == "" {
				return "no-rds", nil
			}
			dp, err := terrak8s.RunKubectlAndGetOutputE(t, mdbNs, "get", RdsInstanceKind, rdsName, "-o", "jsonpath={.spec.forProvider.deletionProtection}")
			if err != nil {
				return "", err
			}
			if dp != "false" {
				return "", fmt.Errorf("deletionProtection=%q", dp)
			}
			return "propagated", nil
		})
}
