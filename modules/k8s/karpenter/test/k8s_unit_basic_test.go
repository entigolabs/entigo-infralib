package test

import (
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sKarpenterAWSBiz(t *testing.T) {
	testK8sKarpenter(t, "aws", "biz")
}

func TestK8sKarpenterAWSPri(t *testing.T) {
	testK8sKarpenter(t, "aws", "pri")
}

func testK8sKarpenter(t *testing.T, cloudName string, envName string) {
  	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)
	
	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, namespaceName, 30, 10*time.Second)
	if err != nil {
		t.Fatal("Karpenter deployment error", err)
	}
}
