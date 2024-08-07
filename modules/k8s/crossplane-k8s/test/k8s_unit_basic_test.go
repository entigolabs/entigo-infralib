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
	terraaws "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/helm"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestK8sCrossplaneAWSBiz(t *testing.T) {
	testK8sCrossplane(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "./k8s_unit_basic_test_aws_biz.yaml", "runner-main-biz", "aws")
}

func TestK8sCrossplaneAWSPri(t *testing.T) {
	testK8sCrossplane(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "./k8s_unit_basic_test_aws_pri.yaml", "runner-main-pri", "aws")
}

// func TestK8sCrossplaneGKEBiz(t *testing.T) {
// 	testK8sCrossplane(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "./k8s_unit_basic_test_gke_biz.yaml", "runner-main-biz", "google")
// }

// func TestK8sCrossplaneGKEPri(t *testing.T) {
// 	testK8sCrossplane(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "./k8s_unit_basic_test_gke_pri.yaml", "runner-main-pri", "google")
// }

func testK8sCrossplane(t *testing.T, contextName, valuesFile, runnerName, cloudName string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := fmt.Sprintf("crossplane-system")
	releaseName := "crossplane-system"

	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	setValues["installProvider"] = "false"
	setValues["installProviderConfig"] = "false"

	if prefix != "runner-main" {
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}

	switch cloudName {
	case "aws":
		awsRegion := terraaws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
		iamrole := terraaws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/%s/iam_role", runnerName))
		setValues["awsRole"] = iamrole

	case "google":

	}

	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)

	helmOptions := &helm.Options{
		ValuesFiles:       []string{fmt.Sprintf("../values-%s.yaml", cloudName), valuesFile},
		SetValues:         setValues,
		KubectlOptions:    kubectlOptions,
		BuildDependencies: false,
		ExtraArgs:         extraArgs,
	}

	if os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
		defer helm.Delete(t, helmOptions, releaseName, true)
		// terrak8s.DeleteNamespace(t, kubectlOptions, namespaceName)
	}

	// Install K8s DeploymentRuntimeConfig
	helmOptions.SetValues = setValues
	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)

	_, err = k8s.WaitUntilDeploymentRuntimeConfigAvailable(t, kubectlOptions, fmt.Sprintf("k8s-%s", releaseName), 60, 1*time.Second)
	require.NoError(t, err, "DeploymentRuntimeConfigAvailable error")

	// Install K8s provider
	setValues["installProvider"] = "true"
	helmOptions.SetValues = setValues
	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)

	k8sprovider, k8serr := k8s.WaitUntilProviderAvailable(t, kubectlOptions, fmt.Sprintf("k8s-%s", releaseName), 60, 1*time.Second)
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
	_, err = k8s.WaitUntilProviderConfigAvailable(t, kubectlOptions, resource, fmt.Sprintf("k8s-%s", releaseName), 60, 1*time.Second)
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
