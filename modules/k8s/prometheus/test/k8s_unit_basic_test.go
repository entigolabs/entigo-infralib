package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gruntwork-io/terratest/modules/helm"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

func TestK8sPrometheusAWSBiz(t *testing.T) {
	testK8sPrometheus(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "biz", "runner-main-biz-int.infralib.entigo.io", "aws")
}

func TestK8sPrometheusAWSPri(t *testing.T) {
	testK8sPrometheus(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "pri", "runner-main-pri.infralib.entigo.io", "aws")
}

func TestK8sPrometheusGKEBiz(t *testing.T) {
	testK8sPrometheus(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "biz", "runner-main-biz-int.gcp.infralib.entigo.io", "google")
}

func TestK8sPrometheusGKEPri(t *testing.T) {
	testK8sPrometheus(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "pri", "runner-main-pri.gcp.infralib.entigo.io", "google")
}

func testK8sPrometheus(t *testing.T, contextName, envName, hostName, cloudName string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := "prometheus"
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	if prefix != "runner-main" {
		namespaceName = fmt.Sprintf("prometheus-%s-%s", envName, prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}
	releaseName := namespaceName
	setValues["prometheus.server.ingress.hosts[0]"] = fmt.Sprintf("%s.%s", releaseName, hostName)

	switch cloudName {
	case "google":
		setValues["google.certificateMap"] = strings.ReplaceAll(hostName, ".", "-")
	}

	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)

	helmOptions := &helm.Options{
		ValuesFiles:       []string{fmt.Sprintf("../values-%s.yaml", cloudName)},
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
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-server", releaseName), 10, 10*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-server deployment error:", releaseName), err)
	}

	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-kube-state-metrics", releaseName), 10, 10*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-kube-state-metrics deployment error:", releaseName), err)
	}

	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, fmt.Sprintf("%s-alertmanager-0", releaseName), 10, 10*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-alertmanager-0 pod error:", releaseName), err)
	}
}
