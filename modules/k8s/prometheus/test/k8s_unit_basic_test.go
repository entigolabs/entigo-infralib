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

func TestK8sPrometheusAWSBiz(t *testing.T) {
	testK8sPrometheus(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "biz", "k8s_unit_basic_test_aws_biz.yaml", "runner-main-biz-int.infralib.entigo.io", "aws")
}

func TestK8sPrometheusAWSPri(t *testing.T) {
	testK8sPrometheus(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "pri", "k8s_unit_basic_test_aws_pri.yaml", "runner-main-pri.infralib.entigo.io", "aws")
}

func TestK8sPrometheusGoogleBiz(t *testing.T) {
	testK8sPrometheus(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "biz", "k8s_unit_basic_test_google_biz.yaml", "runner-main-biz-int.gcp.infralib.entigo.io", "google")
}

func TestK8sPrometheusGooglePri(t *testing.T) {
	testK8sPrometheus(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "pri", "k8s_unit_basic_test_google_biz.yaml", "runner-main-pri.gcp.infralib.entigo.io", "google")
}

func testK8sPrometheus(t *testing.T, contextName, envName, valuesFile, hostName, cloudProvider string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := fmt.Sprintf("prometheus-%s", envName)
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	if prefix != "runner-main" {
		namespaceName = fmt.Sprintf("prometheus-%s-%s", envName, prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}

	releaseName := namespaceName
	gatewayName := ""
	gatewayNamespace := ""

	setValues["global.prefix"] = fmt.Sprintf("%s-%s", prefix, envName)

	switch cloudProvider {
	case "aws":
		setValues["prometheus.server.ingress.hosts[0]"] = fmt.Sprintf("%s.%s", releaseName, hostName)
		gatewayName = fmt.Sprintf("%s-server", namespaceName)

	case "google":
		gatewayNamespace = "google-gateway"

		setValues["global.google.hostname"] = fmt.Sprintf("%s.%s", releaseName, hostName)
		setValues["global.google.gateway.namespace"] = gatewayNamespace

		switch envName {
		case "biz":
			gatewayName = "google-gateway-internal"
		case "pri":
			gatewayName = "google-gateway-external"
		}
		setValues["global.google.gateway.name"] = gatewayName
	}

	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)

	helmOptions := &helm.Options{
		ValuesFiles:       []string{fmt.Sprintf("../values-%s.yaml", cloudProvider), valuesFile},
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
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-server", releaseName), 20, 6*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-server deployment error:", releaseName), err)
	}

	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-kube-state-metrics", releaseName), 20, 6*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-kube-state-metrics deployment error:", releaseName), err)
	}

	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, fmt.Sprintf("%s-alertmanager-0", releaseName), 20, 6*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-alertmanager-0 pod error:", releaseName), err)
	}

	successResponseCode := "301"
	targetURL := fmt.Sprintf("http://%s.%s", releaseName, hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, 100, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "prometheus ingress/gateway test error")

	successResponseCode = "200"
	targetURL = fmt.Sprintf("https://%s.%s/graph", releaseName, hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, 100, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "prometheus ingress/gateway test error")
}
