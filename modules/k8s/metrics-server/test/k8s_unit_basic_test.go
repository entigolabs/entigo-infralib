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

func TestK8sMetricsServerAWSBiz(t *testing.T) {
	testK8sMetricsServer(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "./k8s_unit_basic_test_aws_biz.yaml", "runner-main-biz-int.infralib.entigo.io", "aws")
}

func TestK8sMetricsServerAWSPri(t *testing.T) {
	testK8sMetricsServer(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "./k8s_unit_basic_test_aws_pri.yaml", "runner-main-pri.infralib.entigo.io", "aws")
}

// func TestK8sMetricsServerGKEBiz(t *testing.T) {
// 	testK8sMetricsServer(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "./k8s_unit_basic_test_gke_biz.yaml", "runner-main-biz-int.gcp.infralib.entigo.io", "google")
// }

// func TestK8sMetricsServerGKEPri(t *testing.T) {
// 	testK8sMetricsServer(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "./k8s_unit_basic_test_gke_pri.yaml", "runner-main-pri.gcp.infralib.entigo.io", "google")
// }

func testK8sMetricsServer(t *testing.T, contextName, valuesFile, hostName, cloudName string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	releaseName := "metrics-server"
	namespaceName := "kube-system"

	if prefix != "runner-main" {
		releaseName = fmt.Sprintf("metrics-server-%s", prefix)
		namespaceName = fmt.Sprintf("metrics-server-%s", prefix)
		setValues["metrics-server.apiService.create"] = "false"
		setValues["metrics-server.nameOverride"] = releaseName
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}

	switch cloudName {
	case "google":
		if prefix != "runner-main" {
			setValues["createResourceQuota"] = "true"
		}
	}

	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)

	helmOptions := &helm.Options{
		ValuesFiles:       []string{fmt.Sprintf("../values-%s.yaml", cloudName), valuesFile},
		SetValues:         setValues,
		KubectlOptions:    kubectlOptions,
		BuildDependencies: false,
		ExtraArgs:         extraArgs,
	}

	err = terrak8s.CreateNamespaceE(t, kubectlOptions, namespaceName)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println("Namespace already exists.")
		} else {
			t.Fatal("Error:", err)
		}
	}

	if os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
		defer helm.Delete(t, helmOptions, releaseName, true)
	}

	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)

	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, releaseName, 20, 6*time.Second)
	if err != nil {
		t.Fatal("metric-server deployment error:", err)
	}
}
