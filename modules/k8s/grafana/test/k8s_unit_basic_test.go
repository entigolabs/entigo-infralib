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
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/helm"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

func TestK8sGrafanaAWSBiz(t *testing.T) {
	testK8sGrafana(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "./k8s_unit_basic_test_aws_biz.yaml", "biz", "runner-main-biz-int.infralib.entigo.io", "aws")
}

func TestK8sGrafanaAWSPri(t *testing.T) {
	testK8sGrafana(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "./k8s_unit_basic_test_aws_pri.yaml", "pri", "runner-main-pri.infralib.entigo.io", "aws")
}

func TestK8sGrafanaGoogleBiz(t *testing.T) {
	testK8sGrafana(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "./k8s_unit_basic_test_google_biz.yaml", "biz", "runner-main-biz-int.gcp.infralib.entigo.io", "google")
}

func TestK8sGrafanaGooglePri(t *testing.T) {
	testK8sGrafana(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "./k8s_unit_basic_test_google_pri.yaml", "pri", "runner-main-pri.gcp.infralib.entigo.io", "google")
}

func testK8sGrafana(t *testing.T, contextName, valuesFile, envName, hostname, cloudProvider string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := fmt.Sprintf("grafana-%s", envName)
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	if prefix != "runner-main" {
		namespaceName = fmt.Sprintf("grafana-%s-%s", envName, prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}

	releaseName := namespaceName
	gatewayName := ""
	gatewayNamespace := ""

	setValues["grafana.\"grafana\\.ini\".server.root_url"] = fmt.Sprintf("https://%s.%s", releaseName, hostname)
	setValues["global.prefix"] = fmt.Sprintf("%s-%s", prefix, envName)

	switch cloudProvider {
	case "aws":
		awsRegion := aws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
		awsAccount := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/account", envName))
		clusteroidc := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/oidc_provider", envName))
		region := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/region", envName))
		setValues["global.aws.region"] = region
		setValues["global.aws.account"] = awsAccount
		setValues["global.aws.clusterOIDC"] = clusteroidc
		setValues["grafana.ingress.hosts[0]"] = fmt.Sprintf("%s.%s", releaseName, hostname)
		gatewayName = "grafana"

	case "google":
		projectID := strings.ToLower(os.Getenv("GOOGLE_PROJECT"))
		prometheusHostname := commonGoogle.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-%s-prometheus_hostname/versions/latest", projectID, envName))
		lokiHostname := commonGoogle.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-%s-loki_hostname/versions/latest", projectID, envName))
		gatewayNamespace = "google-gateway"

		switch envName {
		case "biz":
			gatewayName = "google-gateway-internal"
		case "pri":
			gatewayName = "google-gateway-external"
		}

		grafanaServiceAccountName := truncateString(namespaceName, 28)

		setValues["global.loki.hostname"] = lokiHostname
		setValues["global.prometheus.hostname"] = prometheusHostname
		setValues["grafana.serviceAccount.name"] = grafanaServiceAccountName
		setValues["global.google.hostname"] = fmt.Sprintf("%s.%s", releaseName, hostname)
		setValues["global.google.gateway.namespace"] = gatewayNamespace
		setValues["global.google.gateway.name"] = gatewayName
		setValues["global.google.projectID"] = projectID
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
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "grafana", 20, 6*time.Second)
	if err != nil {
		t.Fatal("grafana deployment error:", err)
	}

	retries := 100
	if cloudProvider == "google" && prefix == "runner-main" {
		retries = 300
	}

	successResponseCode := "301"
	targetURL := fmt.Sprintf("http://%s.%s", releaseName, hostname)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "grafana ingress/gateway test error")

	successResponseCode = "200"
	targetURL = fmt.Sprintf("https://%s.%s/login", releaseName, hostname)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "grafana ingress/gateway test error")
}

func truncateString(input string, maxLength int) string {
	if len(input) > maxLength {
		return input[:maxLength]
	}
	return input
}
