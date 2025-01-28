package test

import (
	"testing"
	"time"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/require"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sHelloWorldAWSBiz(t *testing.T) {
	testK8sHelloWorld(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks")
}

func TestK8sHelloWorldAWSPri(t *testing.T) {
	testK8sHelloWorld(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks")
}

func TestK8sHelloWorldGoogleBiz(t *testing.T) {
	testK8sHelloWorld(t, "gke_entigo-infralib2_europe-north1_biz-infra-gke")
}

func TestK8sHelloWorldGooglePri(t *testing.T) {
	testK8sHelloWorld(t, "gke_entigo-infralib2_europe-north1_pri-infra-gke")
}

func testK8sHelloWorld(t *testing.T, contextName string) {
	t.Parallel()
	namespaceName := k8s.GetNamespaceName(t)
        kubectlOptions := k8s.CheckKubectlConnection(t, contextName, namespaceName)

	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, namespaceName, 10, 6*time.Second)
	require.NoError(t, err, "%s deployment %s error: %s",namespaceName, namespaceName, err)

}

