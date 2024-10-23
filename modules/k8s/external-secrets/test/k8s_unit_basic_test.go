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

func TestK8sExternalSecretsAWSBiz(t *testing.T) {
	testK8sExternalSecrets(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "biz", "aws")
}

func TestK8sExternalSecretsAWSPri(t *testing.T) {
	testK8sExternalSecrets(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "pri", "aws")
}

func TestK8sExternalSecretsGoogleBiz(t *testing.T) {
	testK8sExternalSecrets(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "biz", "google")
}

func TestK8sExternalSecretsGooglePri(t *testing.T) {
	testK8sExternalSecrets(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "pri", "google")
}

func testK8sExternalSecrets(t *testing.T, contextName string, envName string, cloudProvider string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := fmt.Sprintf("external-secrets-%s", envName)
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	if prefix != "runner-main" {
		namespaceName = fmt.Sprintf("external-secrets-%s-%s", envName, prefix)
		setValues["external-secrets.installCRDs"] = "false"
		setValues["external-secrets.webhook.create"] = "false"
		setValues["external-secrets.certController.create"] = "false"
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}

	releaseName := namespaceName
	setValues["global.createClusterSecretStore"] = "true"

	switch cloudProvider {
	case "aws":
		awsRegion := aws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
		awsAccount := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/account", envName))
		clusteroidc := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/oidc_provider", envName))
		region := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/region", envName))

		setValues["external-secrets.env[0].value"] = region
		setValues["external-secrets.env[0].name"] = "AWS_DEFAULT_REGION"
		setValues["global.aws.account"] = awsAccount
		setValues["global.aws.clusterOIDC"] = clusteroidc
		setValues["global.aws.region"] = awsRegion

	case "google":
		projectID := os.Getenv("GOOGLE_PROJECT")
		region := os.Getenv("GOOGLE_REGION")
		clusterName := commonGoogle.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-%s-cluster_name/versions/latest", projectID, envName))

		setValues["global.google.projectID"] = projectID
		setValues["global.google.region"] = region
		setValues["global.google.cluster.name"] = clusterName
	}

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
		t.Fatal("external-secrets deployment error:", err)
	}
	if prefix == "runner-main" {
		err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-webhook", namespaceName), 10, 12*time.Second)
		if err != nil {
			t.Fatal("external-secrets-webhook deployment error:", err)
		}

		err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-cert-controller", namespaceName), 10, 12*time.Second)
		if err != nil {
			t.Fatal("external-secrets-cert-controller deployment error:", err)
		}
	}
}
