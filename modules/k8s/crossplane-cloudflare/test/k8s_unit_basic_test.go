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

func TestK8sCrossplaneCloudflareBiz(t *testing.T) {
	testK8sCrossplaneCloudflare(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks", "biz")
}



func testK8sCrossplaneCloudflare(t *testing.T, contextName string, envName string) {
	t.Parallel()

	namespaceName := "crossplane-system"
	releaseName := "crossplane-aws"
	
	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)
	output, err := terrak8s.RunKubectlAndGetOutputE(t, kubectlOptions, "auth", "can-i", "get", "pods")
	require.NoError(t, err, "Unable to connect to context %s cluster %s", contextName, err)
	require.Equal(t, output, "yes")
	
	_, err = k8s.WaitUntilDeploymentRuntimeConfigAvailable(t, kubectlOptions, releaseName, 60, 1*time.Second)
	require.NoError(t, err, "DeploymentRuntimeConfigAvailable error")


	err = k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "upjet-cloudflare.m.upbound.io/v1beta1", []string{"providerconfigs"}, 60, 1*time.Second)
	require.NoError(t, err, "Providerconfigs crd error")
	resource := schema.GroupVersionResource{Group: "upjet-cloudflare.m.upbound.io", Version: "v1beta1", Resource: "providerconfigs"}
	_, err = k8s.WaitUntilProviderConfigAvailable(t, kubectlOptions, resource, releaseName, 60, 1*time.Second)
	require.NoError(t, err, "Provider config error")


}
