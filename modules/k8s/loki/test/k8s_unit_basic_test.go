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
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/helm"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

func TestK8sLokiAWSBiz(t *testing.T) {
	testK8sLoki(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "biz", "k8s_unit_basic_test_aws_biz.yaml", "runner-main-biz-int.infralib.entigo.io", "aws")
}

func TestK8sLokiAWSPri(t *testing.T) {
	testK8sLoki(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "pri", "k8s_unit_basic_test_aws_pri.yaml", "runner-main-pri.infralib.entigo.io", "aws")
}

func TestK8sLokiGoogleBiz(t *testing.T) {
	testK8sLoki(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "biz", "k8s_unit_basic_test_google_biz.yaml", "runner-main-biz-int.gcp.infralib.entigo.io", "google")
}

func TestK8sLokiGooglePri(t *testing.T) {
	testK8sLoki(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "pri", "k8s_unit_basic_test_google_pri.yaml", "runner-main-pri.gcp.infralib.entigo.io", "google")
}

func testK8sLoki(t *testing.T, contextName, envName, valuesFile, hostName, cloudProvider string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	googleProjectID := strings.ToLower(os.Getenv("GOOGLE_PROJECT"))

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := fmt.Sprintf("loki-%s", envName)
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	if prefix != "runner-main" {
		namespaceName = fmt.Sprintf("loki-%s-%s", envName, prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
		setValues["loki.rbac.namespaced"] = "true"
	}

	releaseName := namespaceName
	gatewayName := ""
	gatewayNamespace := ""
	bucketName := fmt.Sprintf("%s-logs", namespaceName)
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
		setValues["global.bucketName"] = bucketName

		setValues["loki.loki.storage.s3.region"] = region
		setValues["loki.loki.storage.bucketNames.chunks"] = bucketName
		setValues["loki.loki.storage.bucketNames.ruler"] = bucketName
		setValues["loki.loki.storage.bucketNames.admin"] = bucketName
		setValues["loki.loki.storage_config.aws.region"] = region
		setValues["loki.loki.storage_config.aws.bucketnames"] = bucketName

		setValues["loki.gateway.ingress.hosts[0].host"] = fmt.Sprintf("%s.%s", releaseName, hostName)
		setValues["loki.gateway.ingress.hosts[0].paths[0].path"] = "/"
		setValues["loki.gateway.ingress.hosts[0].paths[0].pathType"] = "Prefix"

		gatewayName = "loki-gateway"

		if envName == "biz" {
			telemetry_alias_arn := aws.GetParameter(t, awsRegion, "/entigo-infralib/runner-main-biz/telemetry_alias_arn")
			setValues["global.aws.kmsKeyId"] = telemetry_alias_arn
		}

	case "google":
		gatewayNamespace = "google-gateway"

		setValues["loki.loki.storage.bucketNames.chunks"] = bucketName
		setValues["loki.loki.storage.bucketNames.ruler"] = bucketName
		setValues["loki.loki.storage.bucketNames.admin"] = bucketName

		setValues["global.bucketName"] = bucketName

		setValues["global.google.hostname"] = fmt.Sprintf("%s.%s", releaseName, hostName)
		setValues["global.google.projectID"] = googleProjectID
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
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "loki-gateway", 20, 6*time.Second)
	if err != nil {
		t.Fatal("loki-gateway deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "loki-read", 20, 6*time.Second)
	if err != nil {
		t.Fatal("loki-read deployment error:", err)
	}
	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, "loki-write-0", 20, 6*time.Second)
	if err != nil {
		t.Fatal("loki-write-0 pod error:", err)
	}

	successResponseCode := "301"
	targetURL := fmt.Sprintf("http://%s.%s", releaseName, hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, 100, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "loki ingress/gateway test error")

	successResponseCode = "200"
	targetURL = fmt.Sprintf("https://%s.%s", releaseName, hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, 100, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudProvider)
	require.NoError(t, err, "loki ingress/gateway test error")
}
