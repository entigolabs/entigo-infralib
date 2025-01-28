package test

import (
	"testing"
	"time"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sExternalDnsAWSBiz(t *testing.T) {
	testK8sExternalDns(t, "aws", "biz")
}

func TestK8sExternalDnsAWSPri(t *testing.T) {
	testK8sExternalDns(t, "aws", "pri")
}

func TestK8sExternalDnsGoogleBiz(t *testing.T) {
	testK8sExternalDns(t, "google", "biz")
}

func TestK8sExternalDnsGooglePri(t *testing.T) {
	testK8sExternalDns(t, "google", "pri")
}

func testK8sExternalDns(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)
  
	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, namespaceName, 10, 6*time.Second)
	if err != nil {
		t.Fatal("external-dns deployment error:", err)
	}
}
