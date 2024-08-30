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

func TestK8sHarborAWSBiz(t *testing.T) {
	testK8sHarbor(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "biz", "k8s_unit_basic_test_aws_biz.yaml", "runner-main-biz-int.infralib.entigo.io", "aws")
}

func TestK8sHarborAWSPri(t *testing.T) {
	testK8sHarbor(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "pri", "k8s_unit_basic_test_aws_pri.yaml", "runner-main-pri.infralib.entigo.io", "aws")
}

func TestK8sHarborGoogleBiz(t *testing.T) {
	testK8sHarbor(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "biz", "k8s_unit_basic_test_google_biz.yaml", "runner-main-biz-int.gcp.infralib.entigo.io", "google")
}

func TestK8sHarborGooglePri(t *testing.T) {
	testK8sHarbor(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "pri", "k8s_unit_basic_test_google_pri.yaml", "runner-main-pri.gcp.infralib.entigo.io", "google")
}

func testK8sHarbor(t *testing.T, contextName, envName, valuesFile, hostName, cloudProvider string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	googleProjectID := strings.ToLower(os.Getenv("GOOGLE_PROJECT"))
	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := "harbor"
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	if prefix != "runner-main" {
		namespaceName = fmt.Sprintf("harbor-%s", prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}

	releaseName := namespaceName
	gatewayName := ""
	gatewayNamespace := ""

	switch cloudProvider {
	case "aws":
		setValues["harbor.expose.ingress.hosts.core"] = fmt.Sprintf("%s.%s", releaseName, hostName)
		gatewayName = fmt.Sprintf("%s-ingress", namespaceName)

	case "google":
		gatewayNamespace = "google-gateway"

		setValues["namespaceName"] = namespaceName
		setValues["google.projectID"] = googleProjectID
		setValues["google.hostname"] = fmt.Sprintf("%s.%s", releaseName, hostName)
		setValues["google.gateway.namespace"] = gatewayNamespace
		setValues["harbor.persistence.imageChartStorage.gcs.bucket"] = fmt.Sprintf("%s-%s", releaseName, envName)

		switch envName {
		case "biz":
			gatewayName = "google-gateway-internal"
		case "pri":
			gatewayName = "google-gateway-external"
		}
		setValues["google.gateway.name"] = gatewayName
	}

	setValues["harbor.expose.clusterIP.name"] = releaseName
	setValues["harbor.externalURL"] = fmt.Sprintf("https://%s.%s", releaseName, hostName)
	setValues["harbor.harborAdminPassword"] = "Harbor12345"

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

	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, fmt.Sprintf("%s-database-0", releaseName), 20, 6*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-database-0 pod error:", releaseName), err)
	}

	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, fmt.Sprintf("%s-redis-0", releaseName), 20, 6*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-redis-0 pod error:", releaseName), err)
	}

	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, fmt.Sprintf("%s-trivy-0", releaseName), 20, 6*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-trivy-0 pod error:", releaseName), err)
	}

	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-core", releaseName), 20, 6*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-core deployment error:", releaseName), err)
	}

	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-portal", releaseName), 20, 6*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-portal deployment error:", releaseName), err)
	}

	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-registry", releaseName), 30, 10*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-registry deployment error:", releaseName), err)
	}

	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-jobservice", releaseName), 30, 10*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-jobservice deployment error:", releaseName), err)
	}

	successResponseCode := "301"
	targetURL := fmt.Sprintf("http://%s.%s", releaseName, hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, 50, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "harbor ingress/gateway test error")

	successResponseCode = "200"
	targetURL = fmt.Sprintf("https://%s.%s/api/v2.0/ping", releaseName, hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, 50, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "harbor ingress/gateway test error")
}
