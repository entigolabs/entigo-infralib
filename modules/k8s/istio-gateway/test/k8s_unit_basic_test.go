package test

import (
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sIstioGatewayAWSBiz(t *testing.T) {
	testK8sIstioGateway(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks", "biz", "biz-net-route53.infralib.entigo.io")
}

func TestK8sIstioGatewayAWSPri(t *testing.T) {
	testK8sIstioGateway(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks", "pri", "pri-net-route53.infralib.entigo.io")
}

func TestK8sIstioGatewayGoogleBiz(t *testing.T) {
	testK8sIstioGateway(t, "gke_entigo-infralib2_europe-north1_biz-infra-gke", "biz", "biz-net-dns.gcp.infralib.entigo.io")
}

func TestK8sIstioGatewayGooglePri(t *testing.T) {
	testK8sIstioGateway(t, "gke_entigo-infralib2_europe-north1_pri-infra-gke", "pri", "pri-net-dns.gcp.infralib.entigo.io")
}

func testK8sIstioGateway(t *testing.T, contextName string, envName string, hostName string) {
  	t.Parallel()
	namespaceName := "istio-gateway"
        kubectlOptions := k8s.CheckKubectlConnection(t, contextName, namespaceName)

	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "istio-gateway", 10, 5*time.Second)
	if err != nil {
		t.Fatal("istio-gateway deployment error:", err)
	}
}
