package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	commonGoogle "github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/gruntwork-io/terratest/modules/helm"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

func TestK8sGoogleGatewayBiz(t *testing.T) {
	testK8sGoogleGateway(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "biz")
}

func TestK8sGoogleGatewayPri(t *testing.T) {
	testK8sGoogleGateway(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "pri")
}

func testK8sGoogleGateway(t *testing.T, contextName, envName string) {
	t.Parallel()
	spew.Dump("")

	projectID := os.Getenv("GOOGLE_PROJECT")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := fmt.Sprintf("google-gateway")
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	if prefix != "runner-main" {
		namespaceName = fmt.Sprintf("google-gateway-%s-%s", prefix, envName)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}

	releaseName := namespaceName

	switch envName {
	case "pri":
		externalCertificateMap := commonGoogle.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-pri-int_zone_id/versions/latest", projectID))
		setValues["global.google.externalCertificateMap"] = externalCertificateMap
		setValues["global.createInternal"] = "false"
		setValues["global.createExternal"] = "true"
	case "biz":
		externalCertificateMap := commonGoogle.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-biz-pub_zone_id/versions/latest", projectID))
		internalCertificateMap := commonGoogle.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-biz-int_zone_id/versions/latest", projectID))
		setValues["global.google.externalCertificateMap"] = externalCertificateMap
		setValues["global.google.internalCertificateMap"] = internalCertificateMap
		setValues["global.createInternal"] = "true"
		setValues["global.createExternal"] = "true"
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

	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)
	_, err = k8s.WaitUntilK8SGatewayAvailable(t, kubectlOptions, fmt.Sprintf("%s-external", releaseName), 50, 6*time.Second)
	require.NoError(t, err, "google-gateway not available error")

	switch envName {
	case "biz":
		_, err = k8s.WaitUntilK8SGatewayAvailable(t, kubectlOptions, fmt.Sprintf("%s-internal", releaseName), 50, 6*time.Second)
		require.NoError(t, err, "google-gateway not available error")
	}
}
