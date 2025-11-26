package test

import (
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestK8sCrossplaneAWSBiz(t *testing.T) {
	testK8sCrossplaneAWS(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks", "biz")
}

func TestK8sCrossplaneAWSPri(t *testing.T) {
	testK8sCrossplaneAWS(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks", "pri")
}

func testK8sCrossplaneAWS(t *testing.T, contextName string, envName string) {
	t.Parallel()

	namespaceName := "crossplane-system"
	releaseName := "crossplane-sql"

	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)
	output, err := terrak8s.RunKubectlAndGetOutputE(t, kubectlOptions, "auth", "can-i", "get", "pods")
	require.NoError(t, err, "Unable to connect to context %s cluster %s", contextName, err)
	require.Equal(t, output, "yes")

	_, err = k8s.WaitUntilDeploymentRuntimeConfigAvailable(t, kubectlOptions, releaseName, 60, 1*time.Second)
	require.NoError(t, err, "DeploymentRuntimeConfigAvailable error")

	// Install AWS provider
	provider, err := k8s.WaitUntilProviderAvailable(t, kubectlOptions, "contrib-provider-sql", 60, 1*time.Second)
	require.NoError(t, err, "Provider aws error")
	assert.NotNil(t, provider, "Provider aws is nil")
	providerDeployment := k8s.GetStringValue(provider.Object, "status", "currentRevision")
	assert.NotEmpty(t, providerDeployment, "Provider aws currentRevision is empty")
	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, providerDeployment, 60, 1*time.Second)
}
