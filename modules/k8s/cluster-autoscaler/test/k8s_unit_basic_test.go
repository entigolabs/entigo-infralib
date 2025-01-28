package test

import (
	"fmt"
	"testing"
	"time"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/require"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sClusterAutoscalerAWSBiz(t *testing.T) {
	testK8sClusterAutoscaler(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks", "biz")
}

func TestK8sClusterAutoscalerAWSPri(t *testing.T) {
	testK8sClusterAutoscaler(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks", "pri")
}


func testK8sClusterAutoscaler(t *testing.T, contextName string, envName string) {
	t.Parallel()
	namespaceName := fmt.Sprintf("cluster-autoscaler-%s", envName)
        kubectlOptions := k8s.CheckKubectlConnection(t, contextName, namespaceName)
	
	
	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-aws-cluster-autoscaler", namespaceName), 50, 6*time.Second)
	require.NoError(t, err, "aws-cluster-autoscaler deployment %s error: %s", namespaceName, err)

}
