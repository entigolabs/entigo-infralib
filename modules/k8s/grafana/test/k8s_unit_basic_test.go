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
	testK8sGrafana(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks", "biz", "biz-net-route53.infralib.entigo.io", "aws")
}

func TestK8sGrafanaAWSPri(t *testing.T) {
	testK8sGrafana(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks", "pri", "pri-net-route53.infralib.entigo.io", "aws")
}

func TestK8sGrafanaGoogleBiz(t *testing.T) {
	testK8sGrafana(t, "gke_entigo-infralib2_europe-north1_biz-infra-gke", "biz", "biz-net-dns.gcp.infralib.entigo.io", "google")
}

func TestK8sGrafanaGooglePri(t *testing.T) {
	testK8sGrafana(t, "gke_entigo-infralib2_europe-north1_pri-infra-gke", "pri", "pri-net-dns.gcp.infralib.entigo.io", "google")
}

func testK8sGrafana(t *testing.T, contextName, envName, hostname, cloudProvider string) {
  	t.Parallel()
	namespaceName := fmt.Sprintf("grafana-%s", envName)
        kubectlOptions := k8s.CheckKubectlConnection(t, contextName, namespaceName)
  
  
	gatewayName := ""
	gatewayNamespace := ""

	switch cloudProvider {
	case "aws":
		gatewayName = "grafana"

	case "google":
		gatewayNamespace = "google-gateway"
		gatewayName = "google-gateway-external"
	}

	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "grafana", 20, 6*time.Second)
	if err != nil {
		t.Fatal("grafana deployment error:", err)
	}

	retries := 100

	successResponseCode := "301"
	targetURL := fmt.Sprintf("http://grafana.%s", hostname)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "grafana ingress/gateway test error")

	successResponseCode = "200"
	targetURL = fmt.Sprintf("https://grafana.%s/login", hostname)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "grafana ingress/gateway test error")
}

