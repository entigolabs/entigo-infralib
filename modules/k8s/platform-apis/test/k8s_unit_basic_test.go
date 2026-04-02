package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
)

func TestK8sPlatformApisAWSBiz(t *testing.T) {
	testPlatformApis(t, "aws", "biz")
}

// func TestK8sPlatformApisAWSPri(t *testing.T) {
// 	testPlatformApis(t, "aws", "pri")
// }

func testPlatformApis(t *testing.T, cloudName, envName string) {
	t.Parallel()

	kubectlOptions, _ := k8s.CheckKubectlConnection(t, cloudName, envName)
	cfg := loadSuiteConfig()

	cluster, argocd := setupKubectlClients(kubectlOptions, envName)

	waitPackagesReady(t, cfg, cluster)
	if t.Failed() {
		return
	}

	t.Run("setup-zone", func(t *testing.T) {
		setupZoneSync(t, cluster, argocd)
	})
	if t.Failed() {
		t.Fatal("Zones deployment failed. Can not run tests.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("parallel-tests", func(t *testing.T) {
		runSuite := func(name string, fn func(*testing.T, context.Context)) {
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				defer func() {
					if t.Failed() {
						cancel()
					}
				}()
				fn(t, ctx)
			})
		}

		if cfg.Has("zone") {
			runSuite("zone", func(t *testing.T, ctx context.Context) { testZone(t, ctx, cluster) })
		}
		if cfg.Has("postgresql") {
			runSuite("postgresql", func(t *testing.T, ctx context.Context) { testPostgresql(t, ctx, cluster, argocd) })
		}
		if cfg.Has("valkey") {
			runSuite("valkey", func(t *testing.T, ctx context.Context) { testValkey(t, ctx, cluster, argocd) })
		}
		if cfg.Has("cronjob") {
			runSuite("cronjob", func(t *testing.T, ctx context.Context) { testCronjob(t, ctx, cluster, argocd) })
		}
		//TODO: Kafka tests placeholder. Currently disabled as kafka function work in progress
		/*if cfg.Has("kafka") {
			runSuite("kafka", func(t *testing.T, ctx context.Context) { testKafka(t, ctx, cluster, argocd) })
		}*/
		if cfg.Has("repository") {
			runSuite("repository", func(t *testing.T, ctx context.Context) { testRepository(t, ctx, cluster, argocd) })
		}
		if cfg.Has("s3bucket") {
			runSuite("s3bucket", func(t *testing.T, ctx context.Context) { testS3Bucket(t, ctx, cluster, argocd) })
		}
		if cfg.Has("webapp") {
			runSuite("webapp", func(t *testing.T, ctx context.Context) { testWebApp(t, ctx, cluster, argocd) })
		}
		if cfg.Has("webaccess") {
			runSuite("webaccess", func(t *testing.T, ctx context.Context) { testWebAccess(t, ctx, cluster, argocd) })
		}
	})
}

// waitPackagesReady waits for the Crossplane packages (Functions + Configurations) for active suites.
func waitPackagesReady(t *testing.T, cfg SuiteConfig, cluster *terrak8s.KubectlOptions) {
	t.Helper()
	t.Run("packages", func(t *testing.T) {
		if cfg.Has("cronjob") {
			checkPlatformApisHaveRequiredPackages(t, cluster, CronjobConfigurationName, WorkloadFunctionName)
		}
		//  TODO: Update when kafka in go ready
		/*if cfg.Has("kafka") {
			t.Run("kafka-configuration", func(t *testing.T) {
				t.Parallel()
				waitCrossplanePackageReady(t, cluster, ConfigurationKind, KafkaConfigurationName)
			})
		}*/
		if cfg.Has("postgresql") {
			checkPlatformApisHaveRequiredPackages(t, cluster, PostgresqlConfigurationName, DatabaseFunctionName)
		}
		if cfg.Has("repository") {
			checkPlatformApisHaveRequiredPackages(t, cluster, RepositoryConfigurationName, ArtifactFunctionName)
		}
		if cfg.Has("s3bucket") {
			checkPlatformApisHaveRequiredPackages(t, cluster, S3BucketConfigurationName, StorageFunctionName)
		}
		if cfg.Has("valkey") {
			checkPlatformApisHaveRequiredPackages(t, cluster, ValkeyConfigurationName, DatabaseFunctionName)
		}
		if cfg.Has("webaccess") {
			checkPlatformApisHaveRequiredPackages(t, cluster, WebaccessConfigurationName, NetwokringFunctionName)
		}
		if cfg.Has("webapp") {
			checkPlatformApisHaveRequiredPackages(t, cluster, WebappConfigurationName, WorkloadFunctionName)
		}
		if cfg.Has("zone") {
			checkPlatformApisHaveRequiredPackages(t, cluster, ZoneConfigurationName, TenancyFunctionName)
		}
	})
}

func checkPlatformApisHaveRequiredPackages(t *testing.T, cluster *terrak8s.KubectlOptions, configurationName, functionName string) {
	t.Run("configuration", func(t *testing.T) {
		t.Parallel()
		waitCrossplanePackageReady(t, cluster, ConfigurationKind, configurationName)
	})
	t.Run("function", func(t *testing.T) {
		t.Parallel()
		waitCrossplanePackageReady(t, cluster, FunctionKind, functionName)
	})
}

func setupKubectlClients(opts *terrak8s.KubectlOptions, envName string) (*terrak8s.KubectlOptions, *terrak8s.KubectlOptions) {
	argocdNamespace := fmt.Sprintf("argocd-%s", envName)
	cluster := terrak8s.NewKubectlOptions(opts.ContextName, opts.ConfigPath, "")
	argocd := terrak8s.NewKubectlOptions(opts.ContextName, opts.ConfigPath, argocdNamespace)
	return cluster, argocd
}

// setupZoneSync deploys zone
func setupZoneSync(t *testing.T, cluster, argocd *terrak8s.KubectlOptions) {
	t.Helper()

	applyFile(t, cluster, "./templates/zone_test_application.yaml")
	syncWithRetry(t, argocd, ZoneApplicationName)

	for _, zone := range []string{ZoneAName, ZoneBName} {
		waitSyncedAndReady(t, cluster, ZoneKind, zone, 30, 10*time.Second)
		waitZoneNodegroupReady(t, cluster, zone)
	}

	waitApplicationHealthy(t, argocd, ZoneApplicationName)
}

func waitZoneNodegroupReady(t *testing.T, cluster *terrak8s.KubectlOptions, zone string) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("zone %q NodeGroup Ready", zone), 30, 10*time.Second,
		func() (string, error) {
			status, err := terrak8s.RunKubectlAndGetOutputE(t, cluster, "get", NodeGroupKind,
				"-l", fmt.Sprintf("crossplane.io/composite=%s", zone),
				"-o", `jsonpath={.items[0].status.conditions[?(@.type=="Ready")].status}`)
			if err != nil {
				return "", err
			}
			if status != "True" {
				return "", fmt.Errorf("zone %q NodeGroup not Ready: %q", zone, status)
			}
			return status, nil
		})
	require.NoError(t, err, "zone %q NodeGroup never became Ready", zone)
}
