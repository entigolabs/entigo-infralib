package test

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTerraformBasicBiz(t *testing.T) {
	testTerraformBasic(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz")
}

func TestTerraformBasicPri(t *testing.T) {
	testTerraformBasic(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri")
}

func testTerraformBasic(t *testing.T, contextName string) {
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := fmt.Sprintf("crossplane-system")
	releaseName := "crossplane"

	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	if prefix != "runner-main" {
		//releaseName = fmt.Sprintf("crossplane-%s", prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}

	kubectlOptions := k8s.NewKubectlOptions(contextName, "", namespaceName)

	setValues["installProvider"] = "false"
	setValues["installProviderConfig"] = "false"
	setValues["installBucket"] = "false"
	helmOptions := &helm.Options{
		SetValues:         setValues,
		KubectlOptions:    kubectlOptions,
		BuildDependencies: false,
		ExtraArgs:         extraArgs,
	}

	if os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
		defer helm.Delete(t, helmOptions, releaseName, true)
		//k8s.DeleteNamespace(t, kubectlOptions, namespaceName)
	}

	err = k8s.CreateNamespaceE(t, kubectlOptions, namespaceName)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println("Namespace already exists.")
		} else {
			t.Fatal("Error:", err)
		}
	}

	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)
	err = k8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "crossplane", 60, 1*time.Second)
	if err != nil {
		t.Fatal("Crossplane deployment error:", err)
	}
	err = k8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "crossplane-rbac-manager", 60, 1*time.Second)
	if err != nil {
		t.Fatal("Crossplane-rbac-manager deployment error:", err)
	}
	err = WaitUntilResourcesAvailable(t, kubectlOptions, "pkg.crossplane.io/v1", []string{"providers"}, 60, 1*time.Second)
	if err != nil {
		t.Fatal("Providers crd error:", err)
	}
	err = WaitUntilResourcesAvailable(t, kubectlOptions, "pkg.crossplane.io/v1alpha1", []string{"controllerconfigs"}, 60, 1*time.Second)
	if err != nil {
		t.Fatal("Controllerconfigs crd error:", err)
	}

	setValues["installProvider"] = "true"
	helmOptions.SetValues = setValues
	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)

	provider, err := WaitUntilProviderAvailable(t, kubectlOptions, "aws-crossplane", 60, 1*time.Second)
	if err != nil {
		t.Fatal("Provider error:", err)
	}
	assert.NotNil(t, provider, "Provider is nil")
	providerDeployment := GetStringValue(provider.Object, "status", "currentRevision")
	assert.NotEmpty(t, providerDeployment, "Provider currentRevision is empty")
	err = k8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, providerDeployment, 60, 1*time.Second)
	if err != nil {
		t.Fatalf("Provider deployment %s error: %s", providerDeployment, err)
	}
	_, err = WaitUntilControllerConfigAvailable(t, kubectlOptions, "aws-crossplane", 60, 1*time.Second)
	if err != nil {
		t.Fatal("Controller config error:", err)
	}

	setValues["installProviderConfig"] = "true"
	helmOptions.SetValues = setValues
	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)

	err = WaitUntilResourcesAvailable(t, kubectlOptions, "aws.crossplane.io/v1beta1", []string{"providerconfigs"}, 60, 1*time.Second)
	if err != nil {
		t.Fatal("Providerconfigs crd error:", err)
	}
	_, err = WaitUntilProviderConfigAvailable(t, kubectlOptions, "aws-crossplane", 60, 1*time.Second)
	if err != nil {
		t.Fatal("Provider config error:", err)
	}

	setValues["installBucket"] = "true"
	bucketName := "entigo-infralib-test" + strings.ToLower(random.UniqueId()) + "-" + releaseName
	setValues["bucketName"] = bucketName
	helmOptions.SetValues = setValues
	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)

	err = WaitUntilBucketAvailable(t, "eu-north-1", bucketName, 60, 1*time.Second)
	if err != nil {
		t.Fatal("Creating bucket error:", err)
	}
}
