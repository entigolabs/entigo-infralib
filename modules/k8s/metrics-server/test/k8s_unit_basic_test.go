package test

import (
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sMetricsServerAWSBiz(t *testing.T) {
	testK8sMetricsServer(t, "aws", "biz")
}

func TestK8sMetricsServerAWSPri(t *testing.T) {
	testK8sMetricsServer(t, "aws", "pri")
}

func testK8sMetricsServer(t *testing.T, cloudName string, envName string) {
  	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)


	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, namespaceName, 20, 6*time.Second)
	if err != nil {
		t.Fatal("metric-server deployment error:", err)
	}
}
