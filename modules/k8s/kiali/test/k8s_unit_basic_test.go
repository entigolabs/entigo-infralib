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

func TestK8sKialiAWSBiz(t *testing.T) {
	testK8sKiali(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "biz", "k8s_unit_basic_test_aws_biz.yaml", "runner-main-biz-int.infralib.entigo.io", "aws")
}

func TestK8sKialiAWSPri(t *testing.T) {
	testK8sKiali(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "pri", "k8s_unit_basic_test_aws_pri.yaml", "runner-main-pri.infralib.entigo.io", "aws")
}

func TestK8sKialiGKEBiz(t *testing.T) {
	testK8sKiali(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "biz", "k8s_unit_basic_test_gke_biz.yaml", "runner-main-biz-int.gcp.infralib.entigo.io", "google")
}

func TestK8sKialiGKEPri(t *testing.T) {
	testK8sKiali(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "pri", "k8s_unit_basic_test_gke_pri.yaml", "runner-main-pri.gcp.infralib.entigo.io", "google")
}

func testK8sKiali(t *testing.T, contextName, envName, valuesFile, hostName, cloudName string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := fmt.Sprintf("kiali-%s", envName)
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	if prefix != "runner-main" {
		namespaceName = fmt.Sprintf("kiali-%s-%s", envName, prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}

	releaseName := namespaceName
	gatewayName := ""
	gatewayNamespace := ""

	setValues["kiali-server.fullnameOverride"] = namespaceName
	setValues["kiali-server.server.web_fqdn"] = fmt.Sprintf("%s.%s", releaseName, hostName)

	switch cloudName {
	case "google":
		gatewayNamespace = "gcp-gateway"
		setValues["google.hostname"] = fmt.Sprintf("%s.%s", releaseName, hostName)
		setValues["google.gateway.namespace"] = gatewayNamespace
		switch envName {
		case "biz":
			gatewayName = "gcp-gateway-internal"
		case "pri":
			gatewayName = "gcp-gateway-external"
		}
		setValues["google.gateway.name"] = gatewayName
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
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, namespaceName, 20, 6*time.Second)
	if err != nil {
		t.Fatal("kiali deployment error:", err)
	}

	successResponseCode := "301"
	targetURL := fmt.Sprintf("http://%s.%s/kiali", releaseName, hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, 100, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudName)
	require.NoError(t, err, "kiali ingress/gateway test error")

	successResponseCode = "200"
	targetURL = fmt.Sprintf("https://%s.%s/kiali", releaseName, hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, 100, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudName)
	require.NoError(t, err, "kiali ingress/gateway test error")
}
