package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sKubeStateMetricsAWSBiz(t *testing.T) {
	testK8sKubeStateMetrics(t, "aws", "biz")
}

func TestK8sKubeStateMetricsAWSPri(t *testing.T) {
	testK8sKubeStateMetrics(t, "aws", "pri")
}

func TestK8sKubeStateMetricsGoogleBiz(t *testing.T) {
	testK8sKubeStateMetrics(t, "google", "biz")
}

func TestK8sKubeStateMetricsGooglePri(t *testing.T) {
	testK8sKubeStateMetrics(t, "google", "pri")
}

func testK8sKubeStateMetrics(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)

	// fullnameOverride pins the Deployment name to "kube-state-metrics".
	deploymentName := "kube-state-metrics"

	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, deploymentName, 20, 6*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s %s deployment error:", namespaceName, deploymentName), err)
	}
}
