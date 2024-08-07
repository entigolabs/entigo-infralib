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
)

func TestK8sCrossplaneAWSBiz(t *testing.T) {
	testK8sCrossplane(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "./k8s_unit_basic_test_aws_biz.yaml", "aws")
}

func TestK8sCrossplaneAWSPri(t *testing.T) {
	testK8sCrossplane(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "./k8s_unit_basic_test_aws_pri.yaml", "aws")
}

func TestK8sCrossplaneGKEBiz(t *testing.T) {
	testK8sCrossplane(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "./k8s_unit_basic_test_gke_biz.yaml", "google")
}

func TestK8sCrossplaneGKEPri(t *testing.T) {
	testK8sCrossplane(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "./k8s_unit_basic_test_gke_pri.yaml", "google")
}

func testK8sCrossplane(t *testing.T, contextName, valuesFile, cloudName string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := "crossplane-system"
	releaseName := "crossplane-system"

	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	if prefix != "runner-main" {
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
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
	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, "crossplane", 10, 6*time.Second)
	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, "crossplane-rbac-manager", 10, 6*time.Second)

	switch cloudName {
	case "aws":
		err = k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "pkg.crossplane.io/v1", []string{"providers"}, 60, 6*time.Second)
		require.NoError(t, err, "Providers crd error")

		err = k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "pkg.crossplane.io/v1beta1", []string{"deploymentruntimeconfigs"}, 60, 6*time.Second)
		require.NoError(t, err, "DeploymentRuntimeConfig crd error")
	}
}
