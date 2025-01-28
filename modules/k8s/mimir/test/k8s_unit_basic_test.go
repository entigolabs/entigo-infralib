package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/require"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sMimirAWSBiz(t *testing.T) {
	testK8sMimir(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks", "biz", "biz-net-route53-int.infralib.entigo.io", "aws")
}

func TestK8sMimirAWSPri(t *testing.T) {
	testK8sMimir(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks", "pri", "pri-net-route53.infralib.entigo.io", "aws")
}

func TestK8sMimirGoogleBiz(t *testing.T) {
	testK8sMimir(t, "gke_entigo-infralib2_europe-north1_biz-infra-gke", "biz", "biz-net-dns-int.gcp.infralib.entigo.io", "google")
}

func TestK8sMimirGooglePri(t *testing.T) {
	testK8sMimir(t, "gke_entigo-infralib2_europe-north1_pri-infra-gke", "pri", "pri-net-dns.gcp.infralib.entigo.io", "google")
}

func testK8sMimir(t *testing.T, contextName string, envName string, hostName string, cloudProvider string) {
  	t.Parallel()
	namespaceName := fmt.Sprintf("mimir-%s", envName)
        kubectlOptions := k8s.CheckKubectlConnection(t, contextName, namespaceName)

	
	gatewayName := ""
	gatewayNamespace := ""

	switch cloudProvider {
	case "aws":
		gatewayName = "mimir-gateway"

	case "google":
		gatewayNamespace = "google-gateway"

		switch envName {
		case "biz":
			gatewayName = "google-gateway-internal"
		case "pri":
			gatewayName = "google-gateway-external"
		}
	}

	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "mimir-gateway", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-gateway deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "mimir-distributor", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-distributor deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "mimir-querier", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-querier deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "mimir-query-frontend", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-query-frontend deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "mimir-ruler", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-ruler deployment error:", err)
	}
	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, "mimir-compactor-0", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-compactor-0 pod error:", err)
	}
	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, "mimir-ingester-0", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-ingester-0 pod error:", err)
	}
	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, "mimir-store-gateway-0", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-store-gateway-0 pod error:", err)
	}

	retries := 100

	successResponseCode := "301"
	targetURL := fmt.Sprintf("http://mimir.%s", hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "mimir ingress/gateway test error")

	successResponseCode = "200"
	targetURL = fmt.Sprintf("https://mimir.%s", hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "mimir ingress/gateway test error")
}
