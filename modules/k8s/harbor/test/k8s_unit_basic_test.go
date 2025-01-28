package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/require"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sHarborAWSBiz(t *testing.T) {
	testK8sHarbor(t, "aws", "biz")
}

func TestK8sHarborAWSPri(t *testing.T) {
	testK8sHarbor(t, "aws", "pri")
}

func TestK8sHarborGoogleBiz(t *testing.T) {
	testK8sHarbor(t, "google", "biz")
}

func TestK8sHarborGooglePri(t *testing.T) {
	testK8sHarbor(t, "google", "pri")
}

func testK8sHarbor(t *testing.T, cloudName string, envName string) {
  	t.Parallel()
	
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)
	
	gatewayName, gatewayNamespace, hostName := k8s.GetGatewayConfig(t, cloudName, envName, "default")
	
	if cloudName == "aws" {
		gatewayName = fmt.Sprintf("%s-ingress", namespaceName)
	}
	
	err := terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, fmt.Sprintf("%s-database-0", namespaceName), 20, 6*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-database-0 pod error:", namespaceName), err)
	}

	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, fmt.Sprintf("%s-redis-0", namespaceName), 20, 6*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-redis-0 pod error:", namespaceName), err)
	}

	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, fmt.Sprintf("%s-trivy-0", namespaceName), 20, 6*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-trivy-0 pod error:", namespaceName), err)
	}

	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-core", namespaceName), 20, 6*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-core deployment error:", namespaceName), err)
	}

	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-portal", namespaceName), 20, 6*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-portal deployment error:", namespaceName), err)
	}

	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-registry", namespaceName), 30, 10*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-registry deployment error:", namespaceName), err)
	}

	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-jobservice", namespaceName), 30, 10*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-jobservice deployment error:", namespaceName), err)
	}
        retries := 100

	successResponseCode := "301"
	targetURL := fmt.Sprintf("http://%s", hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudName)
	require.NoError(t, err, fmt.Sprintf("%s ingress/gateway test error", namespaceName))

	successResponseCode = "200"
	targetURL = fmt.Sprintf("https://%s/api/v2.0/ping", hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudName)
	require.NoError(t, err, fmt.Sprintf("%s ingress/gateway test error", namespaceName))
}
