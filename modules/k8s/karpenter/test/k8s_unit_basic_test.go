package test

import (
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/require"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sKarpenterAWSBiz(t *testing.T) {
	testK8sKarpenter(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks", "biz", "aws")
}

func TestK8sKarpenterAWSPri(t *testing.T) {
	testK8sKarpenter(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks", "pri", "aws")
}

func testK8sKarpenter(t *testing.T, contextName, envName, cloudProvider string) {
  	t.Parallel()
	namespaceName := fmt.Sprintf("karpenter-%s", envName)
        kubectlOptions := k8s.CheckKubectlConnection(t, contextName, namespaceName)
	
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, namespaceName, 30, 10*time.Second)
	if err != nil {
		t.Fatal("Karpenter deployment error", err)
	}
}
