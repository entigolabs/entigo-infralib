package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	commonGoogle "github.com/entigolabs/entigo-infralib-common/google"
	"github.com/gruntwork-io/terratest/modules/helm"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

func TestK8sPromtailAWSBiz(t *testing.T) {
	testK8sPromtail(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "biz", "runner-main-biz-int.infralib.entigo.io", "aws")
}

func TestK8sPromtailAWSPri(t *testing.T) {
	testK8sPromtail(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "pri", "runner-main-pri.infralib.entigo.io", "aws")
}

func TestK8sPromtailGoogleBiz(t *testing.T) {
	testK8sPromtail(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "biz", "runner-main-biz-int.gcp.infralib.entigo.io", "google")
}

func TestK8sPromtailGooglePri(t *testing.T) {
	testK8sPromtail(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "pri", "runner-main-pri.gcp.infralib.entigo.io", "google")
}

func testK8sPromtail(t *testing.T, contextName, envName, hostName, cloudProvider string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := fmt.Sprintf("promtail-%s", envName)
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	if prefix != "runner-main" {
		namespaceName = fmt.Sprintf("promtail-%s-%s", envName, prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}
	releaseName := namespaceName
	lokiHostname := ""

	switch cloudProvider {
	case "aws":
		lokiHostname = fmt.Sprintf("http://loki-gateway.loki-%s", envName)

	case "google":
		projectID := strings.ToLower(os.Getenv("GOOGLE_PROJECT"))
		lokiHostname = commonGoogle.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-%s-loki_hostname/versions/latest", projectID, envName))
	}

	setValues["promtail.config.clients[0].url"] = fmt.Sprintf("http://%s/loki/api/v1/push", lokiHostname)

	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)

	helmOptions := &helm.Options{
		ValuesFiles:       []string{fmt.Sprintf("../values-%s.yaml", cloudProvider)},
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
}
