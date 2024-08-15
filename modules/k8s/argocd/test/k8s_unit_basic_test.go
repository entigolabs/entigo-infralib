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

func TestK8sArgocdAWSBiz(t *testing.T) {
	testK8sArgocd(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "./k8s_unit_basic_test_aws_biz.yaml", "runner-main-biz-int.infralib.entigo.io", "aws")
}

func TestK8sArgocdAWSPri(t *testing.T) {
	testK8sArgocd(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "./k8s_unit_basic_test_aws_pri.yaml", "runner-main-pri.infralib.entigo.io", "aws")
}

func TestK8sArgocdGKEBiz(t *testing.T) {
	testK8sArgocd(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "./k8s_unit_basic_test_gke_biz.yaml", "runner-main-biz.gcp.infralib.entigo.io", "google")
}

func TestK8sArgocdGKEPri(t *testing.T) {
	testK8sArgocd(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "./k8s_unit_basic_test_gke_pri.yaml", "runner-main-pri.gcp.infralib.entigo.io", "google")
}

func testK8sArgocd(t *testing.T, contextName, valuesFile, hostName, cloudName string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := "argocd"
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	setValues["argocd.global.domain"] = fmt.Sprintf("argocd.%s", hostName)

	if prefix != "runner-main" {
		namespaceName = fmt.Sprintf("argocd-%s", prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
		setValues["argocd.crds.install"] = "false"
		setValues["argocd.global.domain"] = fmt.Sprintf("%s.%s", namespaceName, hostName)
	}

	gatewayName := namespaceName
	releaseName := namespaceName

	switch cloudName {
	case "aws":
		gatewayName = fmt.Sprintf("%s-server", namespaceName)
	case "google":
		setValues["google.certificateMap"] = strings.ReplaceAll(hostName, ".", "-")
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

	successResponseCode := "301"
	targetURL := fmt.Sprintf("http://%s", setValues["argocd.global.domain"])
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, 100, 6*time.Second, gatewayName, namespaceName, targetURL, successResponseCode, cloudName)
	require.NoError(t, err, "argocd ingress/gateway test error")

	successResponseCode = "200"
	targetURL = fmt.Sprintf("https://%s", setValues["argocd.global.domain"])
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, 100, 6*time.Second, gatewayName, namespaceName, targetURL, successResponseCode, cloudName)
	require.NoError(t, err, "argocd ingress/gateway test error")
}
