package test

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
)

const (
	PostgresqlConfigurationName = "platform-apis-postgresql"
	DatabaseFunctionName        = "platform-apis-database-fn"
	PostgresqlNamespaceName     = "test-postgresql"
	PostgresqlApplicationName   = "test-postgresql"
)

func testPlatformApisPostgresql(t *testing.T, clusterOptions *terrak8s.KubectlOptions, zonesReady <-chan struct{}, zonesReadySuccess *atomic.Bool) {
	defer func() {
		if t.Failed() {
			return
		}
		cleanupStart := time.Now()
		cleanupPostgresqlResources(t, clusterOptions)
		fmt.Printf("TIMING: Cleanup took %s\n", time.Since(cleanupStart))
	}()

	waitForZonesReady(t, zonesReady, zonesReadySuccess)

	aAppsOptions := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, AAppsNamespace)
	namespaceOptions := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, PostgresqlNamespaceName)

	testWaitAndSyncPostgresqlApplication(t, aAppsOptions)

	instanceStart := time.Now()
	runPostgresqlInstanceTests(t, namespaceOptions)
	fmt.Printf("TIMING: PostgreSQL instance tests took %s\n", time.Since(instanceStart))

	if t.Failed() {
		return
	}

	userDbStart := time.Now()
	runPostgresqlUserAndDatabaseTests(t, namespaceOptions)
	fmt.Printf("TIMING: User and database tests took %s\n", time.Since(userDbStart))
}

func testWaitAndSyncPostgresqlApplication(t *testing.T, aAppsOptions *terrak8s.KubectlOptions) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for Application '%s' to exist", PostgresqlApplicationName), 30, 10*time.Second, func() (string, error) {
		name, err := terrak8s.RunKubectlAndGetOutputE(t, aAppsOptions, "get", "application", PostgresqlApplicationName, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if name == "" {
			return "", fmt.Errorf("application '%s' not found yet", PostgresqlApplicationName)
		}
		return name, nil
	})
	require.NoError(t, err, fmt.Sprintf("Application '%s' not found in namespace '%s'", PostgresqlApplicationName, AAppsNamespace))

	syncAndWaitApplication(t, aAppsOptions, PostgresqlApplicationName, 60, 10*time.Second)
}
