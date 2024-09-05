package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

func TestK8sIstioGatewayAWSBiz(t *testing.T) {
	testK8sIstioGateway(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "biz", "./k8s_unit_basic_test_aws_biz.yaml", "runner-main-biz-int.infralib.entigo.io", "aws")
}

func TestK8sIstioGatewayAWSPri(t *testing.T) {
	testK8sIstioGateway(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "pri", "./k8s_unit_basic_test_aws_pri.yaml", "runner-main-pri.infralib.entigo.io", "aws")
}

func TestK8sIstioGatewayGoogleBiz(t *testing.T) {
	testK8sIstioGateway(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "biz", "./k8s_unit_basic_test_google_biz.yaml", "runner-main-biz-int.gcp.infralib.entigo.io", "google")
}

func TestK8sIstioGatewayGooglePri(t *testing.T) {
	testK8sIstioGateway(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "pri", "./k8s_unit_basic_test_google_pri.yaml", "runner-main-pri.gcp.infralib.entigo.io", "google")
}

func testK8sIstioGateway(t *testing.T, contextName, envName, valuesFile, hostName, cloudProvider string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := "istio-gateway"
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	if prefix != "runner-main" {
		namespaceName = fmt.Sprintf("istio-gateway-%s", prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}
	releaseName := namespaceName
	gatewayName := ""
	gatewayNamespace := ""

	switch cloudProvider {
	case "aws":
		awsRegion := aws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
		certificateArn := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/pub_cert_arn", envName))
		setValues["global.aws.certificateArn"] = certificateArn

	case "google":
		gatewayNamespace = "google-gateway"
		gatewayName = "google-gateway-external"
		setValues["global.google.hostname"] = hostName
		setValues["global.google.gateway.namespace"] = gatewayNamespace
		setValues["global.google.gateway.name"] = gatewayName
	}

	kubectlOptions := k8s.NewKubectlOptions(contextName, "", namespaceName)

	helmOptions := &helm.Options{
		ValuesFiles:       []string{valuesFile, fmt.Sprintf("../values-%s.yaml", cloudProvider)},
		SetValues:         setValues,
		KubectlOptions:    kubectlOptions,
		BuildDependencies: false,
		ExtraArgs:         extraArgs,
	}

	if os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
		defer helm.Delete(t, helmOptions, releaseName, true)
		// k8s.DeleteNamespace(t, kubectlOptions, namespaceName)
	}

	err = k8s.CreateNamespaceE(t, kubectlOptions, namespaceName)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println("Namespace already exists.")
		} else {
			t.Fatal("Error:", err)
		}
	}

	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)
	err = k8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "istio-gateway", 10, 5*time.Second)
	if err != nil {
		t.Fatal("istio-gateway deployment error:", err)
	}
}
