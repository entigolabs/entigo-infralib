package test

import (
	"fmt"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
)

func TestK8sAlloyAWSBiz(t *testing.T) {
	testK8sAlloy(t, "aws", "biz")
}

func TestK8sAlloyAWSPri(t *testing.T) {
	testK8sAlloy(t, "aws", "pri")
}

func TestK8sAlloyGoogleBiz(t *testing.T) {
	testK8sAlloy(t, "google", "biz")
}

func TestK8sAlloyGooglePri(t *testing.T) {
	testK8sAlloy(t, "google", "pri")
}

func testK8sAlloy(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)

	logsDaemonSetName := fmt.Sprintf("%s-logs", namespaceName)

	daemonSetName, err := terrak8s.GetDaemonSetE(t, kubectlOptions, logsDaemonSetName)
	if err != nil {
		t.Fatal(fmt.Sprintf("Daemonset %s error:", namespaceName), err)
	}
	assert.NotEmpty(t, daemonSetName, "Daemonset was not returned")

	metricsDaemonSetName := fmt.Sprintf("%s-metrics", namespaceName)

	metricsDaemonSet, err := terrak8s.GetDaemonSetE(t, kubectlOptions, metricsDaemonSetName)
	if err != nil {
		t.Fatal(fmt.Sprintf("Daemonset %s error:", metricsDaemonSetName), err)
	}
	assert.NotEmpty(t, metricsDaemonSet, "Metrics daemonset was not returned")

	metricsClusterName := fmt.Sprintf("%s-metricscluster", namespaceName)

	metricsCluster, err := terrak8s.GetDeploymentE(t, kubectlOptions, metricsClusterName)
	if err != nil {
		t.Fatal(fmt.Sprintf("Deployment %s error:", metricsClusterName), err)
	}
	assert.NotEmpty(t, metricsCluster, "Metricscluster deployment was not returned")
}
