package test

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/gruntwork-io/terratest/modules/helm"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestK8sGrafanaBiz(t *testing.T) {
	testK8sGrafana(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "biz", "runner-main-biz-int.infralib.entigo.io")
}

func TestK8sGrafanaPri(t *testing.T) {
	testK8sGrafana(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "pri", "runner-main-pri.infralib.entigo.io")
}

func testK8sGrafana(t *testing.T, contextName string, envName string, hostName string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)
	
	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix")) 
	namespaceName := fmt.Sprintf("grafana-%s", envName)
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)
	
	awsRegion := aws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
	account := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/account",envName))
	clusteroidc := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/oidc_provider",envName))
	region := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/region",envName))
	
	setValues["awsRegion"] = region
	setValues["awsAccount"] = account
	setValues["clusterOIDC"] = clusteroidc

	if prefix != "runner-main" {
	   namespaceName = fmt.Sprintf("grafana-%s-%s", envName, prefix)
	   extraArgs["upgrade"] = []string{"--skip-crds"}
	   extraArgs["install"] = []string{"--skip-crds"}
	}
	releaseName := namespaceName
	setValues["grafana.ingress.hosts[0]"] = fmt.Sprintf("%s.%s", releaseName, hostName)
	setValues["grafana.\"grafana\\.ini\".server.root_url"] = fmt.Sprintf("https://%s.%s", releaseName, hostName)
		
	
	
	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)

	helmOptions := &helm.Options{
		SetValues:         setValues,
		KubectlOptions:    kubectlOptions,
		BuildDependencies: false,
		ExtraArgs:         extraArgs,
	}

	if os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
		defer helm.Delete(t, helmOptions, releaseName, true)
		//terrak8s.DeleteNamespace(t, kubectlOptions, namespaceName)
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
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "grafana", 10, 6*time.Second)
	if err != nil {
		t.Fatal("grafana deployment error:", err)
	}

}