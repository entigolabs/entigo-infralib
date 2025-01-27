package test

import (
	"fmt"
	"testing"
	"time"
	"os"
	"strings"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/require"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sArgocdAWSBiz(t *testing.T) {
	testK8sArgocd(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks", "biz", "biz-net-route53.infralib.entigo.io", "aws")
}

func TestK8sArgocdAWSPri(t *testing.T) {
	testK8sArgocd(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks", "pri", "pri-net-route53.infralib.entigo.io", "aws")
}

func TestK8sArgocdGoogleBiz(t *testing.T) {
	testK8sArgocd(t, "gke_entigo-infralib2_europe-north1_biz-infra-gke", "biz", "biz-net-dns.gcp.infralib.entigo.io", "google")
}

func TestK8sArgocdGooglePri(t *testing.T) {
	testK8sArgocd(t, "gke_entigo-infralib2_europe-north1_pri-infra-gke", "pri", "pri-net-dns.gcp.infralib.entigo.io", "google")
}

func testK8sArgocd(t *testing.T, contextName, envName, hostName, cloudProvider string) {
	t.Parallel()
	namespaceName := fmt.Sprintf("argocd-%s", envName)
        kubectlOptions := k8s.CheckKubectlConnection(t, contextName, namespaceName)
	
	appName := strings.TrimSpace(strings.ToLower(os.Getenv("APP_NAME")))
	
	gatewayName := ""
	gatewayNamespace := ""

	switch cloudProvider {
	case "aws":
		gatewayName = fmt.Sprintf("%s-server", namespaceName)
	case "google":
		gatewayNamespace = "google-gateway"
		gatewayName = "google-gateway-external"
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
	if cloudProvider == "google" && strings.Contains(appName, "runner-main") {
		retries = 300
	}

	successResponseCode := "301"
	targetURL := fmt.Sprintf("http://%s", hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "argocd ingress/gateway test error")

	successResponseCode = "200"
	targetURL = fmt.Sprintf("https://argocd.%s", hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "argocd ingress/gateway test error")
}
