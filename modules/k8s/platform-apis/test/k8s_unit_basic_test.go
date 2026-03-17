package test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

const (
	ConfigurationKind = "configuration.pkg.crossplane.io"
)

func TestK8sPlatformApisAWSBiz(t *testing.T) {
	testK8sPlatformApis(t, "aws", "biz")
}

// func TestK8sPlatformApisAWSPri(t *testing.T) {
// 	testK8sPlatformApis(t, "aws", "pri")
// }

func testK8sPlatformApis(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, _ := k8s.CheckKubectlConnection(t, cloudName, envName)

	argocdNamespace := fmt.Sprintf("argocd-%s", envName)
	argocdOptions := terrak8s.NewKubectlOptions(kubectlOptions.ContextName, kubectlOptions.ConfigPath, argocdNamespace)
	clusterOptions := terrak8s.NewKubectlOptions(kubectlOptions.ContextName, kubectlOptions.ConfigPath, "")

	zonesReady := make(chan struct{})
	var zonesReadySuccess atomic.Bool
	var closeOnce sync.Once
	signalZonesReady := func(success bool) {
		closeOnce.Do(func() {
			if success {
				zonesReadySuccess.Store(true)
			}
			close(zonesReady)
		})
	}

	configStart := time.Now()
	t.Run("configurations", func(t *testing.T) {
		t.Run("zone-configuration", func(t *testing.T) {
			t.Parallel()
			waitCrossplanePackageReady(t, clusterOptions, ConfigurationKind, ZoneConfigurationName)
		})
		t.Run("tenancy-function", func(t *testing.T) {
			t.Parallel()
			waitCrossplanePackageReady(t, clusterOptions, FunctionKind, TenancyFunctionName)
		})
		t.Run("postgresql-configuration", func(t *testing.T) {
			t.Parallel()
			waitCrossplanePackageReady(t, clusterOptions, ConfigurationKind, PostgresqlConfigurationName)
		})
		t.Run("database-function", func(t *testing.T) {
			t.Parallel()
			waitCrossplanePackageReady(t, clusterOptions, FunctionKind, DatabaseFunctionName)
		})
	})
	fmt.Printf("[%s] TIMING: Platform-api config and functions tests took %s\n", argocdNamespace, time.Since(configStart))
	if t.Failed() {
		return
	}

	t.Run("zones", func(t *testing.T) {
		t.Parallel()
		testPlatformApisZone(t, argocdNamespace, clusterOptions, argocdOptions, signalZonesReady)
	})

	t.Run("postgresql", func(t *testing.T) {
		t.Parallel()
		testPlatformApisPostgresql(t, clusterOptions, zonesReady, &zonesReadySuccess)
	})
}

func waitForZonesReady(t *testing.T, zonesReady <-chan struct{}, zonesReadySuccess *atomic.Bool) {
	select {
	case <-zonesReady:
		if !zonesReadySuccess.Load() {
			t.Fatal("zones failed to become ready")
		}
	case <-time.After(40 * time.Minute):
		t.Fatal("timed out waiting for zones to become ready")
	}
}
