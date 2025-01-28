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
	testK8sHarbor(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks", "biz", "biz-net-route53-int.infralib.entigo.io", "aws")
}

func TestK8sHarborAWSPri(t *testing.T) {
	testK8sHarbor(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks", "pri", "pri-net-route53.infralib.entigo.io", "aws")
}

func TestK8sHarborGoogleBiz(t *testing.T) {
	testK8sHarbor(t, "gke_entigo-infralib2_europe-north1_biz-infra-gke", "biz", "biz-net-dns-int.gcp.infralib.entigo.io", "google")
}

func TestK8sHarborGooglePri(t *testing.T) {
	testK8sHarbor(t, "gke_entigo-infralib2_europe-north1_pri-infra-gke", "pri", "pri-net-dns.gcp.infralib.entigo.io", "google")
}

func testK8sHarbor(t *testing.T, contextName string, envName string, hostName string, cloudProvider string) {
  	t.Parallel()
	namespaceName := fmt.Sprintf("harbor-%s", envName)
        kubectlOptions := k8s.CheckKubectlConnection(t, contextName, namespaceName)

	gatewayName := ""
	gatewayNamespace := ""
	
	switch cloudProvider {
	case "aws":
		gatewayName = fmt.Sprintf("%s-ingress", namespaceName)


	case "google":
		gatewayNamespace = "google-gateway"
		switch envName {
		case "biz":
			gatewayName = "google-gateway-internal"
		case "pri":
			gatewayName = "google-gateway-external"
		}
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
	targetURL := fmt.Sprintf("http://harbor.%s", hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "harbor ingress/gateway test error")

	successResponseCode = "200"
	targetURL = fmt.Sprintf("https://harbor.%s/api/v2.0/ping", hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "harbor ingress/gateway test error")
}
