package test

import (
	"testing"
	"time"
	"strings"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sCrossplaneK8sAWSBiz(t *testing.T) {
	testK8sCrossplaneK8s(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks")
}

func TestK8sCrossplaneK8sAWSPri(t *testing.T) {
	testK8sCrossplaneK8s(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks")
}

func TestK8sCrossplaneK8sGoogleBiz(t *testing.T) {
	testK8sCrossplaneK8s(t, "gke_entigo-infralib2_europe-north1_biz-infra-gke")
}

func TestK8sCrossplaneK8sGooglePri(t *testing.T) {
	testK8sCrossplaneK8s(t, "gke_entigo-infralib2_europe-north1_pri-infra-gke")
}

func testK8sCrossplaneK8s(t *testing.T, contextName string) {
	t.Parallel()
	namespaceName := "crossplane-system"
	releaseName := "crossplane-k8s"
        kubectlOptions := k8s.CheckKubectlConnection(t, contextName, namespaceName)

	_, err := k8s.WaitUntilDeploymentRuntimeConfigAvailable(t, kubectlOptions, releaseName, 60, 1*time.Second)
	require.NoError(t, err, "DeploymentRuntimeConfigAvailable error")

	k8sprovider, k8serr := k8s.WaitUntilProviderAvailable(t, kubectlOptions, releaseName, 60, 1*time.Second)
	require.NoError(t, k8serr, "Provider k8s error")
	assert.NotNil(t, k8sprovider, "Provider k8s is nil")
	k8sproviderDeployment := k8s.GetStringValue(k8sprovider.Object, "status", "currentRevision")
	assert.NotEmpty(t, k8sproviderDeployment, "Provider k8s currentRevision is empty")
	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, k8sproviderDeployment, 60, 1*time.Second)

	err = k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "kubernetes.crossplane.io/v1alpha1", []string{"providerconfigs"}, 60, 1*time.Second)
	require.NoError(t, err, "Providerconfigs crd error")
	resource := schema.GroupVersionResource{Group: "kubernetes.crossplane.io", Version: "v1alpha1", Resource: "providerconfigs"}
	_, err = k8s.WaitUntilProviderConfigAvailable(t, kubectlOptions, resource, releaseName, 60, 1*time.Second)
	require.NoError(t, err, "Provider config error")

	// Create object
	serviceName := "entigo-infralib-test" + "-" + strings.ToLower(random.UniqueId()) + "-" + releaseName
	object, err := k8s.CreateK8SObject(t, kubectlOptions, serviceName, "./templates/object.yaml")
	require.NoError(t, err, "Creating object error")
	assert.NotNil(t, object, "Object is nil")
	assert.Equal(t, serviceName, object.GetName(), "Object name is not equal")

	_, err = k8s.WaitUntilK8SObjectAvailable(t, kubectlOptions, serviceName, 30, 6*time.Second)
	if err != nil {
		_ = k8s.DeleteK8SObject(t, kubectlOptions, serviceName)
	}
	require.NoError(t, err, "Object syncing error")
	terrak8s.WaitUntilServiceAvailable(t, kubectlOptions, serviceName, 30, 6*time.Second)

	err = k8s.DeleteK8SObject(t, kubectlOptions, serviceName)
	require.NoError(t, err, "Deleting object error")

	err = k8s.WaitUntilK8SObjectDeleted(t, kubectlOptions, serviceName, 10, 6*time.Second)
	require.NoError(t, err, "Object didn't get deleted")
}

