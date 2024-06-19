package test

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/gruntwork-io/terratest/modules/helm"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestK8sArgocdAWSBiz(t *testing.T) {
	testK8sArgocd(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "./k8s_unit_basic_test_aws_biz.yaml", "runner-main-biz-int.infralib.entigo.io")
}

func TestK8sArgocdAWSPri(t *testing.T) {
	testK8sArgocd(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "./k8s_unit_basic_test_aws_pri.yaml", "runner-main-pri.infralib.entigo.io")
}

func TestK8sArgocdGKEBiz(t *testing.T) {
	testK8sArgocd(t, "gke_entigo-infralib_europe-north1_runner-main-biz", "./k8s_unit_basic_test_gke_biz.yaml", "runner-main-biz-int.gcp.infralib.entigo.io")
}

func TestK8sArgocdGKEPri(t *testing.T) {
	testK8sArgocd(t, "gke_entigo-infralib_europe-north1_runner-main-pri", "./k8s_unit_basic_test_gke_pri.yaml", "runner-main-pri.gcp.infralib.entigo.io")
}

func testK8sArgocd(t *testing.T, contextName string, valuesFile string, hostName string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := fmt.Sprintf("argocd")
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	if prefix != "runner-main" {
		namespaceName = fmt.Sprintf("argocd-%s", prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
		setValues["argocd.crds.install"] = "false"
		setValues["argocd.server.config.url"] = fmt.Sprintf("https://%s.%s", namespaceName, hostName)
		setValues["argocd.server.ingress.hosts[0]"] = fmt.Sprintf("%s.%s", namespaceName, hostName)

	}
	releaseName := namespaceName

	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)

	helmOptions := &helm.Options{
		SetValues:         setValues,
		ValuesFiles:       []string{valuesFile},
		KubectlOptions:    kubectlOptions,
		BuildDependencies: false,
		ExtraArgs:         extraArgs,
	}

	if os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
		defer helm.Delete(t, helmOptions, releaseName, true)
		//terrak8s.DeleteNamespace(t, kubectlOptions, namespaceName)
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

	err = k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "argoproj.io/v1alpha1", []string{"applications"}, 60, 1*time.Second)
	require.NoError(t, err, "Argocd no Applications CRD")
	err = k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "argoproj.io/v1alpha1", []string{"applicationsets"}, 60, 1*time.Second)
	require.NoError(t, err, "Argocd no Applicationsets CRD")

	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-server", namespaceName), 20, 6*time.Second)
	if err != nil {
		t.Fatal("argocd-server deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-repo-server", namespaceName), 20, 6*time.Second)
	if err != nil {
		t.Fatal("argocd-repo-server deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-notifications-controller", namespaceName), 20, 6*time.Second)
	if err != nil {
		t.Fatal("argocd-notifications-controller deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-applicationset-controller", namespaceName), 20, 6*time.Second)
	if err != nil {
		t.Fatal("argocd-applicationset-controller deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-dex-server", namespaceName), 20, 6*time.Second)
	if err != nil {
		t.Fatal("argocd-dex-server deployment error:", err)
	}

}
