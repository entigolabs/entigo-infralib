package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/gruntwork-io/terratest/modules/helm"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestK8sCrossplaneAWSBiz(t *testing.T) {
	testK8sCrossplane(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "runner-main-biz")
}

func TestK8sCrossplaneAWSPri(t *testing.T) {
	testK8sCrossplane(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "runner-main-pri")
}

func TestK8sCrossplaneGKEBiz(t *testing.T) {
	testK8sCrossplane(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "runner-main-biz")
}

func TestK8sCrossplaneGKEPri(t *testing.T) {
	testK8sCrossplane(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "runner-main-pri")
}

func testK8sCrossplane(t *testing.T, contextName, runnerName) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := fmt.Sprintf("crossplane-k8s")
	releaseName := "crossplane-k8s"

	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	setValues["installProviderConfig"] = "false"

	if prefix != "runner-main" {
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}

	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)

	helmOptions := &helm.Options{
		SetValues:         setValues,
		KubectlOptions:    kubectlOptions,
		BuildDependencies: false,
		ExtraArgs:         extraArgs,
	}

	if os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
		defer helm.Delete(t, helmOptions, releaseName, true)
		// terrak8s.DeleteNamespace(t, kubectlOptions, namespaceName)
	}

	// Install K8s DeploymentRuntimeConfig and K8s provider
	helmOptions.SetValues = setValues
	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)

	_, err = k8s.WaitUntilDeploymentRuntimeConfigAvailable(t, kubectlOptions, releaseName, 60, 1*time.Second)
	require.NoError(t, err, "DeploymentRuntimeConfigAvailable error")

	k8sprovider, k8serr := k8s.WaitUntilProviderAvailable(t, kubectlOptions, releaseName, 60, 1*time.Second)
	require.NoError(t, k8serr, "Provider k8s error")
	assert.NotNil(t, k8sprovider, "Provider k8s is nil")
	k8sproviderDeployment := k8s.GetStringValue(k8sprovider.Object, "status", "currentRevision")
	assert.NotEmpty(t, k8sproviderDeployment, "Provider k8s currentRevision is empty")
	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, k8sproviderDeployment, 60, 1*time.Second)

	// Install K8s ProviderConfig
	setValues["installProviderConfig"] = "true"
	helmOptions.SetValues = setValues
	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)

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
