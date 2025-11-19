package test

import (
	"fmt"
	"testing"
	"time"
	"os"
	"strings"
	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	terraaws "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
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
	releaseName := "crossplane-kafka"
	
	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)
	output, err := terrak8s.RunKubectlAndGetOutputE(t, kubectlOptions, "auth", "can-i", "get", "pods")
	require.NoError(t, err, "Unable to connect to context %s cluster %s", contextName, err)
	require.Equal(t, output, "yes")
	
	awsRegion := terraaws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
	
	_, err = k8s.WaitUntilDeploymentRuntimeConfigAvailable(t, kubectlOptions, releaseName, 60, 1*time.Second)
	require.NoError(t, err, "DeploymentRuntimeConfigAvailable error")

	// Install AWS provider
	provider, err := k8s.WaitUntilProviderAvailable(t, kubectlOptions, "upbound-provider-kafka", 60, 1*time.Second)
	require.NoError(t, err, "Provider aws error")
	assert.NotNil(t, provider, "Provider aws is nil")
	providerDeployment := k8s.GetStringValue(provider.Object, "status", "currentRevision")
	assert.NotEmpty(t, providerDeployment, "Provider aws currentRevision is empty")
	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, providerDeployment, 60, 1*time.Second)
}
