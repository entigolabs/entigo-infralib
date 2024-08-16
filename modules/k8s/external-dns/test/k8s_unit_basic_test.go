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
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

func TestK8sExternalDnsAWSBiz(t *testing.T) {
	testK8sExternalDns(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "biz", "aws")
}

func TestK8sExternalDnsAWSPri(t *testing.T) {
	testK8sExternalDns(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "pri", "aws")
}

func TestK8sExternalDnsGoogleBiz(t *testing.T) {
	testK8sExternalDns(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "biz", "google")
}

func TestK8sExternalDnsGooglePri(t *testing.T) {
	testK8sExternalDns(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "pri", "google")
}

func testK8sExternalDns(t *testing.T, contextName string, envName string, cloudProvider string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := fmt.Sprintf("external-dns-%s", envName)
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	switch cloudProvider {
	case "aws":
		awsRegion := aws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
		account := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/account", envName))
		clusteroidc := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/oidc_provider", envName))
		region := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/region", envName))

		setValues["external-dns.env[0].value"] = region
		setValues["external-dns.env[0].name"] = "AWS_DEFAULT_REGION"
		setValues["awsAccount"] = account
		setValues["clusterOIDC"] = clusteroidc

	case "google":
		namespaceName = "external-dns"
		setValues["google.projectID"] = strings.ToLower(os.Getenv("GOOGLE_PROJECT"))
		setValues["managedZone"] = fmt.Sprintf("runner-main-%s-gcp-infralib-entigo-io", envName)

	default:
		t.Fatalf("invalid cloud name: %s", cloudProvider)
	}

	if prefix != "runner-main" {
		namespaceName = fmt.Sprintf("external-dns-%s-%s", envName, prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}

	releaseName := namespaceName

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
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, namespaceName, 10, 6*time.Second)
	if err != nil {
		t.Fatal("external-dns deployment error:", err)
	}
}
