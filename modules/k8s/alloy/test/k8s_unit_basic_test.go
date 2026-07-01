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

	nodeAgentDaemonSetName := fmt.Sprintf("%s-node-agent", namespaceName)

	nodeAgentDaemonSet, err := terrak8s.GetDaemonSetE(t, kubectlOptions, nodeAgentDaemonSetName)
	if err != nil {
		t.Fatal(fmt.Sprintf("Daemonset %s error:", nodeAgentDaemonSetName), err)
	}
	assert.NotEmpty(t, nodeAgentDaemonSet, "Node-agent daemonset was not returned")

	// Metrics are opt-in (they need a Mimir), off by default. The biz envs enable
	// them (see test/*_biz.yaml); pri runs logs-only, so cluster-metrics is absent there.
	if envName == "biz" {
		clusterMetricsName := fmt.Sprintf("%s-cluster-metrics", namespaceName)

		clusterMetrics, err := terrak8s.GetDeploymentE(t, kubectlOptions, clusterMetricsName)
		if err != nil {
			t.Fatal(fmt.Sprintf("Deployment %s error:", clusterMetricsName), err)
		}
		assert.NotEmpty(t, clusterMetrics, "Cluster-metrics deployment was not returned")
	}
}
