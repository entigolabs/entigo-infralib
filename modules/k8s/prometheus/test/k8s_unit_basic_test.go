package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/require"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sPrometheusAWSBiz(t *testing.T) {
	testK8sPrometheus(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks", "biz", "biz-net-route53-int.infralib.entigo.io", "aws")
}

func TestK8sPrometheusAWSPri(t *testing.T) {
	testK8sPrometheus(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks", "pri", "pri-net-route53.infralib.entigo.io", "aws")
}

func TestK8sPrometheusGoogleBiz(t *testing.T) {
	testK8sPrometheus(t, "gke_entigo-infralib2_europe-north1_biz-infra-gke", "biz", "biz-net-dns-int.gcp.infralib.entigo.io", "google")
}

func TestK8sPrometheusGooglePri(t *testing.T) {
	testK8sPrometheus(t, "gke_entigo-infralib2_europe-north1_pri-infra-gke", "pri", "pri-net-dns.gcp.infralib.entigo.io", "google")
}

func testK8sPrometheus(t *testing.T, contextName string, envName string, hostName string, cloudProvider string) {
  	t.Parallel()
	namespaceName := fmt.Sprintf("prometheus-%s", envName)
        kubectlOptions := k8s.CheckKubectlConnection(t, contextName, namespaceName)


	gatewayName := ""
	gatewayNamespace := ""

	switch cloudProvider {
	case "aws":
		gatewayName = fmt.Sprintf("%s-server", namespaceName)

	case "google":
		gatewayNamespace = "google-gateway"

		switch envName {
		case "biz":
			gatewayName = "google-gateway-internal"
		case "pri":
			gatewayName = "google-gateway-external"
		}
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

	retries := 100

	successResponseCode := "301"
	targetURL := fmt.Sprintf("http://prometheus.%s", hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "prometheus ingress/gateway test error")

	successResponseCode = "200"
	targetURL = fmt.Sprintf("https://prometheus.%s/graph", hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "prometheus ingress/gateway test error")
}
