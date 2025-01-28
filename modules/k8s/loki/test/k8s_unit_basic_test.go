package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/require"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sLokiAWSBiz(t *testing.T) {
	testK8sLoki(t, "aws", "biz")
}

func TestK8sLokiAWSPri(t *testing.T) {
	testK8sLoki(t, "aws", "pri")
}

func TestK8sLokiGoogleBiz(t *testing.T) {
	testK8sLoki(t, "google", "biz")
}

func TestK8sLokiGooglePri(t *testing.T) {
	testK8sLoki(t, "google", "pri")
}

func testK8sLoki(t *testing.T, cloudName string, envName string) {
  	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)
	
	gatewayName, gatewayNamespace, hostName := k8s.GetGatewayConfig(t, cloudName, envName, "default")
	
	if cloudName == "aws" {
		gatewayName = "loki-gateway"
	}
	

	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "loki-read", 20, 6*time.Second)
	if err != nil {
		t.Fatal("loki-read deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "loki-gateway", 20, 6*time.Second)
	if err != nil {
		t.Fatal("loki-gateway deployment error:", err)
	}
	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, "loki-write-0", 20, 6*time.Second)
	if err != nil {
		t.Fatal("loki-write-0 pod error:", err)
	}
	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, "loki-backend-0", 20, 6*time.Second)
	if err != nil {
		t.Fatal("loki-backend-0 pod error:", err)
	}
	
	retries := 100

	successResponseCode := "301"
	targetURL := fmt.Sprintf("http://%s", hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudName)
	require.NoError(t, err, fmt.Sprintf("%s ingress/gateway test error", namespaceName))

	successResponseCode = "200"
	targetURL = fmt.Sprintf("https://%s", hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudName)
	require.NoError(t, err, fmt.Sprintf("%s ingress/gateway test error", namespaceName))
}
