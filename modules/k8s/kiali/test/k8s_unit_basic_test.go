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
	testK8sKiali(t, "aws", "biz")
}

func TestK8sKialiAWSPri(t *testing.T) {
	testK8sKiali(t, "aws", "pri")
}

func TestK8sKialiGoogleBiz(t *testing.T) {
	testK8sKiali(t, "google", "biz")
}

func TestK8sKialiGooglePri(t *testing.T) {
	testK8sKiali(t, "google", "pri")
}

func testK8sKiali(t *testing.T, cloudName string, envName string) {
  	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)
	
	gatewayName, gatewayNamespace, hostName := k8s.GetGatewayConfig(t, cloudName, envName, "default")
	
	
	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, namespaceName, 30, 10*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("kiali %s deployment error:", namespaceName), err)
	}

	successResponseCode := "200"
	targetURL := fmt.Sprintf("https://%s/kiali", hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, 100, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudName)
	require.NoError(t, err, "kiali ingress/gateway test error")
}
