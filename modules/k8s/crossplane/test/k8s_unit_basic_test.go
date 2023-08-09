package test

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTerraformBasicBiz(t *testing.T) {
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

	kubectlOptions := k8s.NewKubectlOptions("arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "", namespaceName)

	setValues["installProvider"] = "false"
	setValues["installProviderConfig"] = "false"
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
	helmOptionsSecond := &helm.Options{
		SetValues:         setValues,
		KubectlOptions:    kubectlOptions,
		BuildDependencies: false,
		ExtraArgs:         extraArgs,
	}
	helm.Upgrade(t, helmOptionsSecond, helmChartPath, releaseName)
	err = WaitUntilProviderAvailable(t, kubectlOptions, "aws-crossplane", 60, 1*time.Second)
	if err != nil {
		t.Fatal("Provider error:", err)
	}
	// TODO Test if pods have been launched
	//https://entigo.atlassian.net/browse/RD-37
	//Add tests here that check if CRD is created
	time.Sleep(60 * time.Second)
	setValues["installProviderConfig"] = "true"
	helmOptionsThird := &helm.Options{
		SetValues:         setValues,
		KubectlOptions:    kubectlOptions,
		BuildDependencies: false,
		ExtraArgs:         extraArgs,
	}
	helm.Upgrade(t, helmOptionsThird, helmChartPath, releaseName)
}

func TestTerraformBasicPri(t *testing.T) {
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

	kubectlOptions := k8s.NewKubectlOptions("arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "", namespaceName)

	setValues["installProvider"] = "false"
	setValues["installProviderConfig"] = "false"
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
	//https://entigo.atlassian.net/browse/RD-37
	//Add tests here that check if CRD is created
	time.Sleep(60 * time.Second)
	setValues["installProvider"] = "true"
	helmOptionsSecond := &helm.Options{
		SetValues:         setValues,
		KubectlOptions:    kubectlOptions,
		BuildDependencies: false,
		ExtraArgs:         extraArgs,
	}
	helm.Upgrade(t, helmOptionsSecond, helmChartPath, releaseName)
	//https://entigo.atlassian.net/browse/RD-37
	//Add tests here that check if CRD is created
	time.Sleep(60 * time.Second)
	setValues["installProviderConfig"] = "true"
	helmOptionsThird := &helm.Options{
		SetValues:         setValues,
		KubectlOptions:    kubectlOptions,
		BuildDependencies: false,
		ExtraArgs:         extraArgs,
	}
	helm.Upgrade(t, helmOptionsThird, helmChartPath, releaseName)

}
