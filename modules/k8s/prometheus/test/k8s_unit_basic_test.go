package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

func TestK8sPrometheusAWSBiz(t *testing.T) {
	testK8sPrometheus(t, "aws", "biz")
}

func TestK8sPrometheusAWSPri(t *testing.T) {
	testK8sPrometheus(t, "aws", "pri")
}

func TestK8sPrometheusGoogleBiz(t *testing.T) {
	testK8sPrometheus(t, "google", "biz")
}

func TestK8sPrometheusGooglePri(t *testing.T) {
	testK8sPrometheus(t, "google", "pri")
}

func testK8sPrometheus(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)

	gatewayName, gatewayNamespace, hostName, retries := k8s.GetGatewayConfig(t, cloudName, envName, "default")

	if cloudName == "aws" {
		gatewayName = fmt.Sprintf("%s-server", namespaceName)
	}

	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-server", namespaceName), 20, 6*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-server deployment error:", namespaceName), err)
	}

	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-kube-state-metrics", namespaceName), 20, 6*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-kube-state-metrics deployment error:", namespaceName), err)
	}

	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, fmt.Sprintf("%s-alertmanager-0", namespaceName), 20, 6*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-alertmanager-0 pod error:", namespaceName), err)
	}

	successResponseCode := "200"
	targetURL := fmt.Sprintf("https://%s/query", hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudName)
	require.NoError(t, err, fmt.Sprintf("%s ingress/gateway test error", namespaceName))
}
