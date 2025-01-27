package test

import (
	"testing"
	"time"
	"fmt"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sMetricsServerAWSBiz(t *testing.T) {
	testK8sMetricsServer(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks", "biz")
}

func TestK8sMetricsServerAWSPri(t *testing.T) {
	testK8sMetricsServer(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks", "pri")
}

func testK8sMetricsServer(t *testing.T, contextName string, envName string) {
  	t.Parallel()
	namespaceName := fmt.Sprintf("metrics-server-%s", envName)
        kubectlOptions := k8s.CheckKubectlConnection(t, contextName, namespaceName)


	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, namespaceName, 20, 6*time.Second)
	if err != nil {
		t.Fatal("metric-server deployment error:", err)
	}
}
