package test

import (
	"fmt"
	"testing"
	"time"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sExternalDnsAWSBiz(t *testing.T) {
	testK8sExternalDns(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks", "biz")
}

func TestK8sExternalDnsAWSPri(t *testing.T) {
	testK8sExternalDns(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks", "pri")
}

func TestK8sExternalDnsGoogleBiz(t *testing.T) {
	testK8sExternalDns(t, "gke_entigo-infralib2_europe-north1_biz-infra-gke", "biz")
}

func TestK8sExternalDnsGooglePri(t *testing.T) {
	testK8sExternalDns(t, "gke_entigo-infralib2_europe-north1_pri-infra-gke", "pri")
}

func testK8sExternalDns(t *testing.T, contextName string, envName string) {
	t.Parallel()
	namespaceName := fmt.Sprintf("external-dns-%s", envName)
        kubectlOptions := k8s.CheckKubectlConnection(t, contextName, namespaceName)
  
	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, namespaceName, 10, 6*time.Second)
	if err != nil {
		t.Fatal("external-dns deployment error:", err)
	}
}
