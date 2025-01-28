package test

import (
	"testing"
	"time"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/require"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sCrossplaneAWSBiz(t *testing.T) {
	testK8sCrossplane(t, "aws", "biz")
}

func TestK8sCrossplaneAWSPri(t *testing.T) {
	testK8sCrossplane(t, "aws", "pri")
}

func TestK8sCrossplaneGoogleBiz(t *testing.T) {
	testK8sCrossplane(t, "google", "biz")
}

func TestK8sCrossplaneGooglePri(t *testing.T) {
	testK8sCrossplane(t, "google", "pri")
}

func testK8sCrossplane(t *testing.T,  cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, _ := k8s.CheckKubectlConnection(t, cloudName, envName)


	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, "crossplane", 10, 6*time.Second)
	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, "crossplane-rbac-manager", 10, 6*time.Second)

	err := k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "pkg.crossplane.io/v1", []string{"providers"}, 60, 6*time.Second)
	require.NoError(t, err, "Providers crd error")

	err = k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "pkg.crossplane.io/v1beta1", []string{"deploymentruntimeconfigs"}, 60, 6*time.Second)
	require.NoError(t, err, "DeploymentRuntimeConfig crd error")
}
