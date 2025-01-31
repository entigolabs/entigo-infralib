package test

import (
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sIstioGatewayAWSBiz(t *testing.T) {
	testK8sIstioGateway(t, "aws", "biz")
}

func TestK8sIstioGatewayAWSPri(t *testing.T) {
	testK8sIstioGateway(t, "aws", "pri")
}

/* https://entigo.atlassian.net/browse/RD-507
Temporarily disalbed, causes slowdown in httproute processing in GKE
func TestK8sIstioGatewayGoogleBiz(t *testing.T) {
	testK8sIstioGateway(t, "google", "biz")
}

func TestK8sIstioGatewayGooglePri(t *testing.T) {
	testK8sIstioGateway(t, "google", "pri")
}
*/

func testK8sIstioGateway(t *testing.T, cloudName string, envName string) {
  	t.Parallel()
	kubectlOptions, _ := k8s.CheckKubectlConnection(t, cloudName, envName)

	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions,  "istio-gateway", 10, 5*time.Second)
	if err != nil {
		t.Fatal("istio-gateway deployment error:", err)
	}
}
