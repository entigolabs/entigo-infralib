package test

import (
	"fmt"
	"testing"
	"time"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/require"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sHelloWorldBiz(t *testing.T) {
	testK8sHelloWorld(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks", "biz")
}

func TestK8sHelloWorldPri(t *testing.T) {
	testK8sHelloWorld(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks", "pri")
}

func testK8sHelloWorld(t *testing.T, contextName string, envName string) {
	t.Parallel()
	namespaceName := fmt.Sprintf("hello-world-%s", envName)
        kubectlOptions := k8s.CheckKubectlConnection(t, contextName, namespaceName)

	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s", namespaceName), 10, 6*time.Second)
	require.NoError(t, err, "%s deployment %s error: %s",namespaceName, namespaceName, err)

}
