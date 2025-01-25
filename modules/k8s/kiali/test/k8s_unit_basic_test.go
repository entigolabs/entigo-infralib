package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/require"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sKialiAWSBiz(t *testing.T) {
	testK8sKiali(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks", "biz", "runner-main-biz-int.infralib.entigo.io", "aws")
}

func TestK8sKialiAWSPri(t *testing.T) {
	testK8sKiali(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks", "pri", "runner-main-pri.infralib.entigo.io", "aws")
}

func TestK8sKialiGoogleBiz(t *testing.T) {
	testK8sKiali(t, "gke_entigo-infralib2_europe-north1_biz-infra-gke", "biz", "runner-main-biz-int.gcp.infralib.entigo.io", "google")
}

func TestK8sKialiGooglePri(t *testing.T) {
	testK8sKiali(t, "gke_entigo-infralib2_europe-north1_pri-infra-gke", "pri", "runner-main-pri.gcp.infralib.entigo.io", "google")
}

func testK8sKiali(t *testing.T, contextName string, envName string, hostName string, cloudProvider string) {
  	t.Parallel()
	namespaceName := fmt.Sprintf("kiali-%s", envName)
        kubectlOptions := k8s.CheckKubectlConnection(t, contextName, namespaceName)

	gatewayName := ""
	gatewayNamespace := ""

	switch cloudProvider {
	case "aws":
		gatewayName = namespaceName
	case "google":
		gatewayNamespace = "google-gateway"

		switch envName {
		case "biz":
			gatewayName = "google-gateway-internal"
		case "pri":
			gatewayName = "google-gateway-external"
		}
	}
	retries := 100
	
	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "kiali", 30, 10*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("kiali deployment error:", namespaceName), err)
	}

	successResponseCode := "301"
	targetURL := fmt.Sprintf("http://%s.%s/kiali", namespaceName, hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "kiali ingress/gateway test error")

	successResponseCode = "200"
	targetURL = fmt.Sprintf("https://%s.%s/kiali", namespaceName, hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "kiali ingress/gateway test error")
}
