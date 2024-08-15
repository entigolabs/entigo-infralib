package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	commonGCP "github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/gruntwork-io/terratest/modules/helm"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

func TestK8sGCPGatewayBiz(t *testing.T) {
	testK8sGCPGateway(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "biz")
}

// func TestK8sGCPGatewayPri(t *testing.T) {
// 	testK8sGCPGateway(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "pri")
// }

func testK8sGCPGateway(t *testing.T, contextName, envName string) {
	t.Parallel()
	spew.Dump("")

	projectID := os.Getenv("GOOGLE_PROJECT")

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
		externalCertificateMap := commonGCP.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-pri-int_zone_id/versions/latest", projectID))
		setValues["gcp.externalCertificateMap"] = externalCertificateMap
		setValues["createInternal"] = "false"
		setValues["createExternal"] = "true"
	case "biz":
		externalCertificateMap := commonGCP.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-biz-pub_zone_id/versions/latest", projectID))
		internalCertificateMap := commonGCP.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-biz-int_zone_id/versions/latest", projectID))
		setValues["gcp.externalCertificateMap"] = externalCertificateMap
		setValues["gcp.internalCertificateMap"] = internalCertificateMap
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
