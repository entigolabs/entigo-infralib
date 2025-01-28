package test

import (
	"fmt"
	"testing"
	"time"
	//"os"
	//"strings"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/require"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sArgocdAWSBiz(t *testing.T) {
	testK8sArgocd(t, "aws", "biz")
}

func TestK8sArgocdAWSPri(t *testing.T) {
	testK8sArgocd(t, "aws", "pri")
}

func TestK8sArgocdGoogleBiz(t *testing.T) {
	testK8sArgocd(t, "google", "biz")
}

func TestK8sArgocdGooglePri(t *testing.T) {
	testK8sArgocd(t, "google", "pri")
}

func testK8sArgocd(t *testing.T,  cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)
	
	gatewayName, gatewayNamespace, hostName := k8s.GetGatewayConfig(t, cloudName, envName, "external")

	if cloudName == "aws" {
		gatewayName = fmt.Sprintf("%s-server", namespaceName)
	}

	err := k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "argoproj.io/v1alpha1", []string{"applications"}, 60, 1*time.Second)
	require.NoError(t, err, "Argocd no Applications CRD")
	err = k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "argoproj.io/v1alpha1", []string{"applicationsets"}, 60, 1*time.Second)
	require.NoError(t, err, "Argocd no Applicationsets CRD")

	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-server", namespaceName), 20, 6*time.Second)
	if err != nil {
		t.Fatal("argocd-server deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-repo-server", namespaceName), 20, 6*time.Second)
	if err != nil {
		t.Fatal("argocd-repo-server deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-notifications-controller", namespaceName), 20, 6*time.Second)
	if err != nil {
		t.Fatal("argocd-notifications-controller deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-applicationset-controller", namespaceName), 20, 6*time.Second)
	if err != nil {
		t.Fatal("argocd-applicationset-controller deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-dex-server", namespaceName), 20, 6*time.Second)
	if err != nil {
		t.Fatal("argocd-dex-server deployment error:", err)
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
