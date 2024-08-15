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

func TestK8sMimirAWSBiz(t *testing.T) {
	testK8sMimir(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "biz", "k8s_unit_basic_test_aws_biz.yaml", "runner-main-biz-int.infralib.entigo.io", "aws")
}

func TestK8sMimirAWSPri(t *testing.T) {
	testK8sMimir(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "pri", "k8s_unit_basic_test_aws_pri.yaml", "runner-main-pri.infralib.entigo.io", "aws")
}

func TestK8sMimirGKEBiz(t *testing.T) {
	testK8sMimir(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "biz", "k8s_unit_basic_test_gke_biz.yaml", "runner-main-biz-int.gcp.infralib.entigo.io", "google")
}

func TestK8sMimirGKEPri(t *testing.T) {
	testK8sMimir(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "pri", "k8s_unit_basic_test_gke_pri.yaml", "runner-main-pri.gcp.infralib.entigo.io", "google")
}

func testK8sMimir(t *testing.T, contextName, envName, valuesFile, hostName, cloudName string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	googleProjectID := strings.ToLower(os.Getenv("GOOGLE_PROJECT"))

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := fmt.Sprintf("mimir-%s", envName)
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	if prefix != "runner-main" {
		namespaceName = fmt.Sprintf("mimir-%s-%s", envName, prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}
	releaseName := namespaceName
	gatewayName := namespaceName
	bucketName := fmt.Sprintf("%s-logs", namespaceName)
	setValues["global.bucketName"] = bucketName

	switch cloudName {
	case "aws":
		awsRegion := aws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
		account := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/account", envName))
		clusteroidc := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/oidc_provider", envName))
		region := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/region", envName))

		setValues["global.region"] = region
		setValues["global.awsAccount"] = account
		setValues["global.clusterOIDC"] = clusteroidc

		setValues["mimir-distributed.gateway.ingress.hosts[0].host"] = fmt.Sprintf("%s.%s", releaseName, hostName)
		setValues["mimir-distributed.gateway.ingress.hosts[0].paths[0].path"] = "/"
		setValues["mimir-distributed.gateway.ingress.hosts[0].paths[0].pathType"] = "Prefix"

		gatewayName = "mimir-gateway"

	case "google":
		setValues["google.hostname"] = fmt.Sprintf("%s.%s", releaseName, hostName)
		setValues["google.projectID"] = googleProjectID
		setValues["google.certificateMap"] = strings.ReplaceAll(hostName, ".", "-")

		setValues["global.namespaceName"] = namespaceName
	}

	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)

	helmOptions := &helm.Options{
		ValuesFiles:       []string{fmt.Sprintf("../values-%s.yaml", cloudName), valuesFile},
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
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "mimir-gateway", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-gateway deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "mimir-distributor", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-distributor deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "mimir-querier", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-querier deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "mimir-query-frontend", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-query-frontend deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "mimir-ruler", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-ruler deployment error:", err)
	}
	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, "mimir-compactor-0", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-compactor-0 pod error:", err)
	}
	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, "mimir-ingester-0", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-ingester-0 pod error:", err)
	}
	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, "mimir-store-gateway-0", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-store-gateway-0 pod error:", err)
	}

	successResponseCode := "301"
	targetURL := fmt.Sprintf("http://%s.%s", releaseName, hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, 100, 6*time.Second, gatewayName, namespaceName, targetURL, successResponseCode, cloudName)
	require.NoError(t, err, "mimir hostname not available error")

	successResponseCode = "200"
	targetURL = fmt.Sprintf("https://%s.%s", releaseName, hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, 100, 6*time.Second, gatewayName, namespaceName, targetURL, successResponseCode, cloudName)
	require.NoError(t, err, "mimir hostname not available error")
}
