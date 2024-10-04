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

func TestK8sEntigoPortalAgentAWSBiz(t *testing.T) {
	testK8sEntigoPortalAgent(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "biz", "aws")
}

func TestK8sEntigoPortalAgentAWSPri(t *testing.T) {
	testK8sEntigoPortalAgent(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "pri", "aws")
}

//func TestK8sEntigoPortalAgentGoogleBiz(t *testing.T) {
//	testK8sEntigoPortalAgent(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "biz", "google")
//}

//func TestK8sEntigoPortalAgentGooglePri(t *testing.T) {
//	testK8sEntigoPortalAgent(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "pri", "google")
//}

func testK8sEntigoPortalAgent(t *testing.T, contextName, envName, cloudProvider string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	googleProjectID := strings.ToLower(os.Getenv("GOOGLE_PROJECT"))
	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := fmt.Sprintf("portal-agent-%s", envName)
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	if prefix != "runner-main" {
		namespaceName = fmt.Sprintf("portal-agent-%s-%s", envName, prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}

	releaseName := namespaceName

	switch cloudProvider {
	case "aws":
	  
		awsRegion := aws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
		awsAccount := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/account", envName))
		clusterOIDC := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/oidc_provider", envName))
		region := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/region", envName))
		setValues["global.aws.account"] = awsAccount
		setValues["global.aws.clusterOIDC"] = clusterOIDC
		setValues["global.aws.region"] = region

	case "google":
		setValues["global.google.projectID"] = googleProjectID
	}



	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)

	helmOptions := &helm.Options{
		ValuesFiles:       []string{fmt.Sprintf("../values-%s.yaml", cloudProvider), fmt.Sprintf("k8s_unit_basic_test_%s_%s.yaml", cloudProvider,envName)},
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

	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s", releaseName), 30, 10*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s deployment error:", releaseName), err)
	}

}
