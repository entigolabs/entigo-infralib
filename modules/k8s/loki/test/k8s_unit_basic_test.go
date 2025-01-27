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
	testK8sLoki(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks", "biz", "biz-net-route53-int.infralib.entigo.io", "aws")
}

func TestK8sLokiAWSPri(t *testing.T) {
	testK8sLoki(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks", "pri", "pri-net-route53.infralib.entigo.io", "aws")
}

func TestK8sLokiGoogleBiz(t *testing.T) {
	testK8sLoki(t, "gke_entigo-infralib2_europe-north1_biz-infra-gke", "biz", "biz-net-dns-int.gcp.infralib.entigo.io", "google")
}

func TestK8sLokiGooglePri(t *testing.T) {
	testK8sLoki(t, "gke_entigo-infralib2_europe-north1_pri-infra-gke", "pri", "pri-net-dns.gcp.infralib.entigo.io", "google")
}

func testK8sLoki(t *testing.T, contextName string, envName string, hostName string, cloudProvider string) {
  	t.Parallel()
	namespaceName := fmt.Sprintf("loki-%s", envName)
        kubectlOptions := k8s.CheckKubectlConnection(t, contextName, namespaceName)

	gatewayName := ""
	gatewayNamespace := ""

	switch cloudProvider {
	case "aws":

		gatewayName = "loki-gateway"

	case "google":
		gatewayNamespace = "google-gateway"

		switch envName {
		case "biz":
			gatewayName = "google-gateway-internal"
		case "pri":
			gatewayName = "google-gateway-external"
		}
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
	targetURL := fmt.Sprintf("http://loki.%s", hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "loki ingress/gateway test error")

	successResponseCode = "200"
	targetURL = fmt.Sprintf("https://loki.%s", hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "loki ingress/gateway test error")
}
