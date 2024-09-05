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
		awsAccount := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/account", envName))
		clusteroidc := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/oidc_provider", envName))
		region := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/region", envName))

		setValues["external-dns.env[0].value"] = region
		setValues["external-dns.env[0].name"] = "AWS_DEFAULT_REGION"
		setValues["aws.account"] = awsAccount
		setValues["aws.clusterOIDC"] = clusteroidc

		externalZoneID := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/pub_zone_id", envName))
		internalZoneID := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/int_zone_id", envName))
		setValues["external-dns.extraArgs"] = fmt.Sprintf("{--metrics-address=:7979, --zone-id-filter=%s, --zone-id-filter=%s}", externalZoneID, internalZoneID)

	case "google":
		namespaceName = "external-dns"
		projectID := os.Getenv("GOOGLE_PROJECT")
		setValues["google.projectID"] = projectID

		parentZoneID := commonGoogle.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-%s-parent_zone_id/versions/latest", projectID, envName))
		externalZoneID := commonGoogle.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-%s-pub_zone_id/versions/latest", projectID, envName))
		internalZoneID := commonGoogle.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-%s-int_zone_id/versions/latest", projectID, envName))
		setValues["external-dns.extraArgs"] = fmt.Sprintf("{--metrics-address=:7979, --zone-id-filter=%s, --zone-id-filter=%s, --zone-id-filter=%s}", parentZoneID, externalZoneID, internalZoneID)
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
