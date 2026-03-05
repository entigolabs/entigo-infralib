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

	testTestNamespaceCreated(t, clusterOptions)

	namespaceOptions := terrak8s.NewKubectlOptions(clusterOptions.ContextName, clusterOptions.ConfigPath, PostgresqlNamespaceName)

	testApplyAllPostgresqlTemplates(t, namespaceOptions)

	instanceStart := time.Now()
	runPostgresqlInstanceTests(t, namespaceOptions)
	fmt.Printf("TIMING: PostgreSQL instance tests took %s\n", time.Since(instanceStart))

	if t.Failed() {
		return
	}

	//snapshotStart := time.Now()
	//runPostgresqlSnapshotInstanceTests(t, namespaceOptions)
	//fmt.Printf("TIMING: PostgreSQL snapshot instance tests took %s\n", time.Since(snapshotStart))

	//if t.Failed() {
	//	return
	//}

	userDbStart := time.Now()
	runPostgresqlUserAndDatabaseTests(t, namespaceOptions)
	fmt.Printf("TIMING: User and database tests took %s\n", time.Since(userDbStart))
}

func testApplyAllPostgresqlTemplates(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	templates := []string{
		"./templates/postgresql_test_instance.yaml",
		"./templates/postgresql_test_owner_user.yaml",
		"./templates/postgresql_test_user.yaml",
		"./templates/postgresql_test_database_one.yaml",
		"./templates/postgresql_test_database_two.yaml",
		"./templates/postgresql_test_database_minimal.yaml",
	}
	for _, tmpl := range templates {
		_, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "apply", "-f", tmpl, "-n", PostgresqlNamespaceName)
		require.NoError(t, err, fmt.Sprintf("applying %s error", tmpl))
	}
}

func testTestNamespaceCreated(t *testing.T, clusterOptions *terrak8s.KubectlOptions) {
	_, err := retry.DoWithRetryE(t, "create namespace "+PostgresqlNamespaceName, 10, 15*time.Second, func() (string, error) {
		_, createErr := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "create", "namespace", PostgresqlNamespaceName)
		if createErr == nil {
			return "created", nil
		}
		_, getErr := terrak8s.RunKubectlAndGetOutputE(t, clusterOptions, "get", "namespace", PostgresqlNamespaceName)
		if getErr == nil {
			return "already exists", nil
		}
		return "", createErr
	})
	require.NoError(t, err, "namespace "+PostgresqlNamespaceName+" not found")
}
