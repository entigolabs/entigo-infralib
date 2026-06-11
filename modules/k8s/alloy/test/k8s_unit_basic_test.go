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

	logsDaemonSet, err := terrak8s.GetDaemonSetE(t, kubectlOptions, logsDaemonSetName)
	if err != nil {
		t.Fatal(fmt.Sprintf("Daemonset %s error:", logsDaemonSetName), err)
	}
	assert.NotEmpty(t, logsDaemonSet, "Logs daemonset was not returned")

	nodeMetricsDaemonSetName := fmt.Sprintf("%s-node-metrics", namespaceName)

	nodeMetricsDaemonSet, err := terrak8s.GetDaemonSetE(t, kubectlOptions, nodeMetricsDaemonSetName)
	if err != nil {
		t.Fatal(fmt.Sprintf("Daemonset %s error:", nodeMetricsDaemonSetName), err)
	}
	assert.NotEmpty(t, nodeMetricsDaemonSet, "Node-metrics daemonset was not returned")

	clusterMetricsName := fmt.Sprintf("%s-cluster-metrics", namespaceName)

	clusterMetrics, err := terrak8s.GetDeploymentE(t, kubectlOptions, clusterMetricsName)
	if err != nil {
		t.Fatal(fmt.Sprintf("Deployment %s error:", clusterMetricsName), err)
	}
	assert.NotEmpty(t, clusterMetrics, "Cluster-metrics deployment was not returned")
}
