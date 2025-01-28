package test

import (
	"testing"
	"time"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/require"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sHelloWorldAWSBiz(t *testing.T) {
	testK8sHelloWorld(t, "aws", "biz")
}

func TestK8sHelloWorldAWSPri(t *testing.T) {
	testK8sHelloWorld(t, "aws", "pri")
}

func TestK8sHelloWorldGoogleBiz(t *testing.T) {
	testK8sHelloWorld(t, "google", "biz")
}

func TestK8sHelloWorldGooglePri(t *testing.T) {
	testK8sHelloWorld(t, "google", "pri")
}

func testK8sHelloWorld(t *testing.T, cloudName string, envName string) {
	t.Parallel()
        kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)

	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, namespaceName, 10, 6*time.Second)
	require.NoError(t, err, "%s deployment %s error: %s",namespaceName, namespaceName, err)

}

