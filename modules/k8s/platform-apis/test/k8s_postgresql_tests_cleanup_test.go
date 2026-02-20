package test

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
)

func cleanupPostgresqlResources(t *testing.T, clusterOptions *terrak8s.KubectlOptions) {
	pgNsOptions := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, PostgresqlNamespaceName)

	cleanupDeleteForeground(t, pgNsOptions, PostgresqlDatabaseKind, PostgresqlDatabaseName)
	cleanupDeleteForeground(t, pgNsOptions, PostgresqlDatabaseKind, MinimalDatabaseName)
	var wgDbs sync.WaitGroup
	wgDbs.Add(2)
	go func() {
		defer wgDbs.Done()
		cleanupWaitForDeletion(t, pgNsOptions, PostgresqlDatabaseKind, PostgresqlDatabaseName, 30)
	}()
	go func() {
		defer wgDbs.Done()
		cleanupWaitForDeletion(t, pgNsOptions, PostgresqlDatabaseKind, MinimalDatabaseName, 30)
	}()
	wgDbs.Wait()

	cleanupDeleteForeground(t, pgNsOptions, PostgresqlAdminUserKind, PostgresqlRegularUserName)
	cleanupDeleteForeground(t, pgNsOptions, PostgresqlAdminUserKind, PostgresqlAdminUserName)
	var wgUsers sync.WaitGroup
	wgUsers.Add(2)
	go func() {
		defer wgUsers.Done()
		cleanupWaitForDeletion(t, pgNsOptions, PostgresqlAdminUserKind, PostgresqlRegularUserName, 30)
	}()
	go func() {
		defer wgUsers.Done()
		cleanupWaitForDeletion(t, pgNsOptions, PostgresqlAdminUserKind, PostgresqlAdminUserName, 30)
	}()
	wgUsers.Wait()

	cleanupDeleteAllOfKind(t, pgNsOptions, UsageKind)
	cleanupDeleteAllOfKind(t, pgNsOptions, SqlGrantKind)
	cleanupDeleteAllOfKind(t, pgNsOptions, SqlRoleKind)

	cleanupDisableDeletionProtection(t, pgNsOptions)

	cleanupDeleteForeground(t, pgNsOptions, PostgresqlInstanceKind, PostgresqlInstanceName)
	cleanupWaitForDeletion(t, pgNsOptions, PostgresqlInstanceKind, PostgresqlInstanceName, 60)

	cleanupWaitForGeneratedResources(t, pgNsOptions)

	cleanupNamespace(t, pgNsOptions, clusterOptions)
}

func cleanupDeleteForeground(t *testing.T, opts *terrak8s.KubectlOptions, kind string, name string) {
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, opts, "delete", kind, name, "-n", PostgresqlNamespaceName, "--cascade=foreground", "--wait=false", "--ignore-not-found")
}

func cleanupWaitForDeletion(t *testing.T, opts *terrak8s.KubectlOptions, kind string, name string, maxRetries int) {
	_, _ = retry.DoWithRetryE(t, fmt.Sprintf("waiting for %s/%s deletion", kind, name), maxRetries, 10*time.Second, func() (string, error) {
		output, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", kind, name, "-n", PostgresqlNamespaceName, "--ignore-not-found", "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if output != "" {
			return "", fmt.Errorf("%s/%s still exists", kind, name)
		}
		return "deleted", nil
	})
}

func cleanupDeleteAllOfKind(t *testing.T, opts *terrak8s.KubectlOptions, kind string) {
	output, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", kind, "-n", PostgresqlNamespaceName, "-o", "jsonpath={.items[*].metadata.name}", "--ignore-not-found")
	if err != nil || output == "" {
		return
	}
	names := strings.Fields(output)
	for _, name := range names {
		cleanupDeleteForeground(t, opts, kind, name)
	}
	for _, name := range names {
		cleanupWaitForDeletion(t, opts, kind, name, 30)
	}
}

func cleanupDisableDeletionProtection(t *testing.T, opts *terrak8s.KubectlOptions) {
	instanceExists, _ := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", PostgresqlInstanceKind, PostgresqlInstanceName, "-n", PostgresqlNamespaceName, "--ignore-not-found", "-o", "jsonpath={.metadata.name}")
	if instanceExists == "" {
		return
	}

	dp, _ := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", PostgresqlInstanceKind, PostgresqlInstanceName, "-n", PostgresqlNamespaceName, "-o", "jsonpath={.spec.deletionProtection}")
	if dp != "true" {
		return
	}

	_, _ = terrak8s.RunKubectlAndGetOutputE(t, opts, "patch", PostgresqlInstanceKind, PostgresqlInstanceName, "-n", PostgresqlNamespaceName, "--type", "merge", "-p", `{"spec":{"deletionProtection":false}}`)

	_, _ = retry.DoWithRetryE(t, "waiting for RDS deletionProtection=false", 30, 10*time.Second, func() (string, error) {
		rdsName, err := getFirstByLabel(t, opts, RdsInstanceKind, PostgresqlInstanceName)
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
}

func cleanupWaitForGeneratedResources(t *testing.T, opts *terrak8s.KubectlOptions) {
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

	var wg sync.WaitGroup
	for _, gk := range generatedKinds {
		gk := gk
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = retry.DoWithRetryE(t, fmt.Sprintf("waiting for %s deletion", gk.label), 60, 10*time.Second, func() (string, error) {
				output, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", gk.kind, "-l", fmt.Sprintf("crossplane.io/composite=%s", PostgresqlInstanceName), "-o", "jsonpath={.items[*].metadata.name}", "--ignore-not-found")
				if err != nil {
					return "", err
				}
				if output != "" {
					return "", fmt.Errorf("%s still exist: %s", gk.label, output)
				}
				return "deleted", nil
			})
		}()
	}
	wg.Wait()
}

func cleanupNamespace(t *testing.T, pgNsOptions *terrak8s.KubectlOptions, clusterOptions *terrak8s.KubectlOptions) {
	leftovers, _ := terrak8s.RunKubectlAndGetOutputE(t, pgNsOptions, "get", "all", "-n", PostgresqlNamespaceName, "--ignore-not-found", "-o", "name")
	if leftovers != "" {
		_, _ = terrak8s.RunKubectlAndGetOutputE(t, pgNsOptions, "delete", "all", "--all", "-n", PostgresqlNamespaceName, "--cascade=foreground", "--wait=false", "--ignore-not-found")
		time.Sleep(10 * time.Second)
	}
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "delete", "namespace", PostgresqlNamespaceName, "--ignore-not-found", "--wait=true")
}
