package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
)

func cleanupPostgresqlResources(t *testing.T, argocdNamespace string, clusterOptions *terrak8s.KubectlOptions) {
	pgNsOptions := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, PostgresqlNamespaceName)

	// Phase 1: Delete databases (Usage resources ensure Grant is not deleted before Database)
	fmt.Printf("[%s] Cleanup Phase 1: Deleting databases\n", argocdNamespace)
	cleanupDeleteForeground(t, argocdNamespace, pgNsOptions, PostgresqlDatabaseKind, PostgresqlDatabaseName)
	cleanupWaitForDeletion(t, argocdNamespace, pgNsOptions, PostgresqlDatabaseKind, PostgresqlDatabaseName, 30)
	cleanupDeleteForeground(t, argocdNamespace, pgNsOptions, PostgresqlDatabaseKind, MinimalDatabaseName)
	cleanupWaitForDeletion(t, argocdNamespace, pgNsOptions, PostgresqlDatabaseKind, MinimalDatabaseName, 30)

	// Phase 2: Delete users (Roles can now be dropped since databases are gone)
	fmt.Printf("[%s] Cleanup Phase 2: Deleting users\n", argocdNamespace)
	cleanupDeleteForeground(t, argocdNamespace, pgNsOptions, PostgresqlAdminUserKind, PostgresqlRegularUserName)
	cleanupWaitForDeletion(t, argocdNamespace, pgNsOptions, PostgresqlAdminUserKind, PostgresqlRegularUserName, 30)
	cleanupDeleteForeground(t, argocdNamespace, pgNsOptions, PostgresqlAdminUserKind, PostgresqlAdminUserName)
	cleanupWaitForDeletion(t, argocdNamespace, pgNsOptions, PostgresqlAdminUserKind, PostgresqlAdminUserName, 30)

	// Phase 3: Check for leftover Grants, Roles, and Usages, delete if any
	fmt.Printf("[%s] Cleanup Phase 3: Checking for leftover Grants, Roles, and Usages\n", argocdNamespace)
	cleanupDeleteAllOfKind(t, argocdNamespace, pgNsOptions, UsageKind, "Usages")
	cleanupDeleteAllOfKind(t, argocdNamespace, pgNsOptions, SqlGrantKind, "Grants")
	cleanupDeleteAllOfKind(t, argocdNamespace, pgNsOptions, SqlRoleKind, "Roles")

	// Phase 5: Disable deletion protection on PostgreSQLInstance and wait for propagation
	fmt.Printf("[%s] Cleanup Phase 5: Handling deletion protection\n", argocdNamespace)
	cleanupDisableDeletionProtection(t, argocdNamespace, pgNsOptions)

	// Phase 6: Delete PostgreSQLInstance, then immediately patch RDS to skip final snapshot.
	fmt.Printf("[%s] Cleanup Phase 6: Deleting PostgreSQL Instance\n", argocdNamespace)
	cleanupDeleteForeground(t, argocdNamespace, pgNsOptions, PostgresqlInstanceKind, PostgresqlInstanceName)
	cleanupWaitForDeletion(t, argocdNamespace, pgNsOptions, PostgresqlInstanceKind, PostgresqlInstanceName, 60)

	// Phase 7: Verify all instance-generated resources are gone
	fmt.Printf("[%s] Cleanup Phase 7: Verifying generated resources deleted\n", argocdNamespace)
	cleanupWaitForGeneratedResources(t, argocdNamespace, pgNsOptions)

	// Phase 8: Check namespace for leftovers, clean up, delete namespace
	fmt.Printf("[%s] Cleanup Phase 8: Cleaning namespace\n", argocdNamespace)
	cleanupNamespace(t, argocdNamespace, pgNsOptions, clusterOptions)
}

// cleanupDeleteForeground initiates a foreground cascading delete without waiting.
func cleanupDeleteForeground(t *testing.T, argocdNamespace string, opts *terrak8s.KubectlOptions, kind string, name string) {
	fmt.Printf("[%s] Cleanup: deleting %s '%s'\n", argocdNamespace, kind, name)
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, opts, "delete", kind, name, "-n", PostgresqlNamespaceName, "--cascade=foreground", "--wait=false", "--ignore-not-found")
}

// cleanupWaitForDeletion waits until a specific resource no longer exists.
func cleanupWaitForDeletion(t *testing.T, argocdNamespace string, opts *terrak8s.KubectlOptions, kind string, name string, maxRetries int) {
	_, _ = retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for %s '%s' deletion", argocdNamespace, kind, name), maxRetries, 10*time.Second, func() (string, error) {
		output, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", kind, name, "-n", PostgresqlNamespaceName, "--ignore-not-found", "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if output != "" {
			return "", fmt.Errorf("%s '%s' still exists", kind, name)
		}
		return "deleted", nil
	})
	fmt.Printf("[%s] Cleanup: %s '%s' deleted\n", argocdNamespace, kind, name)
}

// cleanupDeleteAllOfKind finds and deletes all resources of a given kind in the namespace.
func cleanupDeleteAllOfKind(t *testing.T, argocdNamespace string, opts *terrak8s.KubectlOptions, kind string, label string) {
	output, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", kind, "-n", PostgresqlNamespaceName, "-o", "jsonpath={.items[*].metadata.name}", "--ignore-not-found")
	if err != nil || output == "" {
		fmt.Printf("[%s] Cleanup: no leftover %s found\n", argocdNamespace, label)
		return
	}
	names := strings.Fields(output)
	fmt.Printf("[%s] Cleanup: found %d leftover %s: %v\n", argocdNamespace, len(names), label, names)
	for _, name := range names {
		cleanupDeleteForeground(t, argocdNamespace, opts, kind, name)
	}
	for _, name := range names {
		cleanupWaitForDeletion(t, argocdNamespace, opts, kind, name, 30)
	}
}

// cleanupDisableDeletionProtection checks and disables deletion protection on the instance and RDS.
func cleanupDisableDeletionProtection(t *testing.T, argocdNamespace string, opts *terrak8s.KubectlOptions) {
	// Check if PostgreSQLInstance exists
	instanceExists, _ := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", PostgresqlInstanceKind, PostgresqlInstanceName, "-n", PostgresqlNamespaceName, "--ignore-not-found", "-o", "jsonpath={.metadata.name}")
	if instanceExists == "" {
		fmt.Printf("[%s] Cleanup: PostgreSQL Instance '%s' not found, skipping deletion protection\n", argocdNamespace, PostgresqlInstanceName)
		return
	}

	// Check current deletion protection status
	dp, _ := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", PostgresqlInstanceKind, PostgresqlInstanceName, "-n", PostgresqlNamespaceName, "-o", "jsonpath={.spec.deletionProtection}")
	if dp == "true" {
		fmt.Printf("[%s] Cleanup: disabling deletionProtection on PostgreSQL Instance '%s'\n", argocdNamespace, PostgresqlInstanceName)
		_, _ = terrak8s.RunKubectlAndGetOutputE(t, opts, "patch", PostgresqlInstanceKind, PostgresqlInstanceName, "-n", PostgresqlNamespaceName, "--type", "merge", "-p", `{"spec":{"deletionProtection":false}}`)

		// Wait for deletion protection to propagate to RDS
		fmt.Printf("[%s] Cleanup: waiting for deletionProtection to propagate to RDS\n", argocdNamespace)
		_, _ = retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for RDS deletionProtection=false", argocdNamespace), 30, 10*time.Second, func() (string, error) {
			rdsName, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", RdsInstanceKind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlInstanceName), "-o", "jsonpath={.items[0].metadata.name}")
			if err != nil || rdsName == "" {
				return "no-rds", nil
			}
			rdsDp, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", RdsInstanceKind, rdsName, "-o", "jsonpath={.spec.forProvider.deletionProtection}")
			if err != nil {
				return "", err
			}
			if rdsDp != "false" {
				return "", fmt.Errorf("RDS deletionProtection is '%s', waiting for 'false'", rdsDp)
			}
			return "propagated", nil
		})
	} else {
		fmt.Printf("[%s] Cleanup: deletionProtection already disabled on PostgreSQL Instance\n", argocdNamespace)
	}

}

// cleanupWaitForGeneratedResources waits for all instance-generated resources to be deleted.
func cleanupWaitForGeneratedResources(t *testing.T, argocdNamespace string, opts *terrak8s.KubectlOptions) {
	generatedKinds := []struct {
		kind  string
		label string
	}{
		{RdsInstanceKind, "RDS Instances"},
		{SecurityGroupRuleKind, "SecurityGroupRules"},
		{SecurityGroupKind, "SecurityGroups"},
		{ExternalSecretKind, "ExternalSecrets"},
		{SqlProviderConfigKind, "ProviderConfigs"},
	}

	for _, gk := range generatedKinds {
		_, _ = retry.DoWithRetryE(t, fmt.Sprintf("[%s] Waiting for %s deletion", argocdNamespace, gk.label), 60, 10*time.Second, func() (string, error) {
			output, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", gk.kind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlInstanceName), "-o", "jsonpath={.items[*].metadata.name}", "--ignore-not-found")
			if err != nil {
				return "", err
			}
			if output != "" {
				return "", fmt.Errorf("%s still exist: %s", gk.label, output)
			}
			return "deleted", nil
		})
		fmt.Printf("[%s] Cleanup: %s deleted\n", argocdNamespace, gk.label)
	}
}

// cleanupNamespace checks for leftover resources and deletes the namespace.
func cleanupNamespace(t *testing.T, argocdNamespace string, pgNsOptions *terrak8s.KubectlOptions, clusterOptions *terrak8s.KubectlOptions) {
	// Check for any leftover resources in the namespace
	leftovers, _ := terrak8s.RunKubectlAndGetOutputE(t, pgNsOptions, "get", "all", "-n", PostgresqlNamespaceName, "--ignore-not-found", "-o", "name")
	if leftovers != "" {
		names := strings.Fields(leftovers)
		fmt.Printf("[%s] Cleanup: found %d leftover resources in namespace: %v\n", argocdNamespace, len(names), names)
		_, _ = terrak8s.RunKubectlAndGetOutputE(t, pgNsOptions, "delete", "all", "--all", "-n", PostgresqlNamespaceName, "--cascade=foreground", "--wait=false", "--ignore-not-found")
		time.Sleep(10 * time.Second)
	} else {
		fmt.Printf("[%s] Cleanup: namespace '%s' is clean\n", argocdNamespace, PostgresqlNamespaceName)
	}

	fmt.Printf("[%s] Cleanup: deleting namespace '%s'\n", argocdNamespace, PostgresqlNamespaceName)
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "delete", "namespace", PostgresqlNamespaceName, "--ignore-not-found", "--wait=true")
	fmt.Printf("[%s] Cleanup: namespace '%s' deleted\n", argocdNamespace, PostgresqlNamespaceName)
}
