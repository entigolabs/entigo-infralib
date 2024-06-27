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
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestK8sCrossplaneBiz(t *testing.T) {
	testK8sCrossplane(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "runner-main-bix")
}

func testK8sCrossplane(t *testing.T, contextName string, runnerName string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := "crossplane-system"
	releaseName := "crossplane"

	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	if prefix != "runner-main" {
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}

	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)

	setValues["installControllerConfig"] = "false"
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
	}

	err = terrak8s.CreateNamespaceE(t, kubectlOptions, namespaceName)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println("Namespace already exists.")
		} else {
			t.Fatal("Error:", err)
		}
	}

	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)
	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, "crossplane", 10, 5*time.Second)
	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, "crossplane-rbac-manager", 10, 5*time.Second)

	_, err = k8s.WaitUntilDeploymentRuntimeConfigAvailable(t, kubectlOptions, "default", 60, 1*time.Second)
	require.NoError(t, err, "DeploymentRuntimeConfigAvailable error")

	setValues["installControllerConfig"] = "true"
	helmOptions.SetValues = setValues
	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)
	_, err = k8s.WaitUntilControllerConfigAvailable(t, kubectlOptions, "my-controller-config", 60, 5*time.Second)
	require.NoError(t, err, "ControllerConfig crd error")

	setValues["installProvider"] = "true"
	helmOptions.SetValues = setValues
	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)
	_, err = k8s.WaitUntilProviderAvailable(t, kubectlOptions, "provider-gcp-storage", 60, 5*time.Second)
	require.NoError(t, err, "Providers crd error")
	_, err = k8s.WaitUntilProviderAvailable(t, kubectlOptions, "upbound-provider-family-gcp", 60, 5*time.Second)
	require.NoError(t, err, "Providers crd error")

	setValues["installProviderConfig"] = "true"
	helmOptions.SetValues = setValues
	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)
	_, err = k8s.WaitUntilProviderConfigAvailable(t, kubectlOptions, schema.GroupVersionResource{Group: "gcp.upbound.io", Version: "v1beta1", Resource: "providerconfigs"}, "workload-id-providerconfig", 60, 5*time.Second)
	require.NoError(t, err, "ProviderConfig crd error")
}
