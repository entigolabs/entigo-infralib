package test

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

const (
	PostgresqlConfigurationName = "platform-apis-postgresql"
	DatabaseFunctionName        = "platform-apis-database-fn"
	PostgresqlNamespaceName     = "test-postgresql"
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

	testTestNamespaceCreated(t, clusterOptions)

	waitForZonesReady(t, zonesReady, zonesReadySuccess)

	namespaceOptions := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, PostgresqlNamespaceName)

	instanceStart := time.Now()
	runPostgresqlInstanceTests(t, namespaceOptions)
	fmt.Printf("TIMING: PostgreSQL instance tests took %s\n", time.Since(instanceStart))

	userDbStart := time.Now()
	runPostgresqlUserAndDatabaseTests(t, namespaceOptions)
	fmt.Printf("TIMING: User and database tests took %s\n", time.Since(userDbStart))
}

func testTestNamespaceCreated(t *testing.T, clusterOptions *terrak8s.KubectlOptions) {
	_, err := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "create", "namespace", PostgresqlNamespaceName)
	if err != nil {
		_, err = terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", "namespace", PostgresqlNamespaceName)
		require.NoError(t, err, "namespace "+PostgresqlNamespaceName+" not found")
	}
}
