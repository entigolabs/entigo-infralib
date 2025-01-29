package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/require"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sGrafanaAWSBiz(t *testing.T) {
	testK8sGrafana(t, "aws", "biz")
}

func TestK8sGrafanaAWSPri(t *testing.T) {
	testK8sGrafana(t, "aws", "pri")
}

func TestK8sGrafanaGoogleBiz(t *testing.T) {
	testK8sGrafana(t, "google", "biz")
}

func TestK8sGrafanaGooglePri(t *testing.T) {
	testK8sGrafana(t, "google", "pri")
}

func testK8sGrafana(t *testing.T, cloudName string, envName string) {
  	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)
	
	gatewayName, gatewayNamespace, hostName := k8s.GetGatewayConfig(t, cloudName, envName, "external")
	
	if cloudName == "aws" {
		gatewayName = "grafana"
	}

	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "grafana", 20, 6*time.Second)
	if err != nil {
		t.Fatal("grafana deployment error:", err)
	}

	successResponseCode = "200"
	targetURL = fmt.Sprintf("https://%s/login", hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, 100, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudName)
	require.NoError(t, err, fmt.Sprintf("%s ingress/gateway test error", namespaceName))
}

