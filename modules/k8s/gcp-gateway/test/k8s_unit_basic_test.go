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

func TestK8sGCPGatewayBiz(t *testing.T) {
	testK8sGCPGateway(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "biz", "runner-main-biz.gcp.infralib.entigo.io", "runner-main-biz-int.gcp.infralib.entigo.io")
}

func TestK8sGCPGatewayPri(t *testing.T) {
	testK8sGCPGateway(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "pri", "runner-main-pri.gcp.infralib.entigo.io", "")
}

func testK8sGCPGateway(t *testing.T, contextName, envName, externalHostname, internalHostname string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := fmt.Sprintf("gcp-gateway")
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	if prefix != "runner-main" {
		namespaceName = fmt.Sprintf("gcp-gateway-%s", prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}

	releaseName := namespaceName

	switch envName {
	case "pri":
		setValues["gcp.externalHostname"] = fmt.Sprintf("%s.%s", releaseName, externalHostname)
		setValues["gcp.externalCertificateMap"] = strings.ReplaceAll(externalHostname, ".", "-")
		setValues["createInternal"] = "false"
		setValues["createExternal"] = "true"
	case "biz":
		setValues["gcp.externalHostname"] = fmt.Sprintf("%s.%s", releaseName, externalHostname)
		setValues["gcp.internalHostname"] = fmt.Sprintf("%s.%s", releaseName, internalHostname)
		setValues["gcp.externalCertificateMap"] = strings.ReplaceAll(externalHostname, ".", "-")
		setValues["gcp.internalCertificateMap"] = strings.ReplaceAll(internalHostname, ".", "-")
		setValues["createInternal"] = "true"
		setValues["createExternal"] = "true"
	}

	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)

	helmOptions := &helm.Options{
		ValuesFiles:       []string{"../values.yaml"},
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

	switch envName {
	case "pri":
		helm.Upgrade(t, helmOptions, helmChartPath, releaseName)
		_, err = k8s.WaitUntilGatewayAvailable(t, kubectlOptions, fmt.Sprintf("%s-external", releaseName), 50, 6*time.Second)
		require.NoError(t, err, "gcp-gateway is not available error")
	case "biz":
		helm.Upgrade(t, helmOptions, helmChartPath, releaseName)
		_, err = k8s.WaitUntilGatewayAvailable(t, kubectlOptions, fmt.Sprintf("%s-external", releaseName), 50, 6*time.Second)
		require.NoError(t, err, "gcp-gateway is not available error")
		_, err = k8s.WaitUntilGatewayAvailable(t, kubectlOptions, fmt.Sprintf("%s-internal", releaseName), 50, 6*time.Second)
		require.NoError(t, err, "gcp-gateway is not available error")
	}
}
