package test

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/gruntwork-io/terratest/modules/helm"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestK8sGrafanaBiz(t *testing.T) {
	testK8sGrafana(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "biz")
}

func TestK8sGrafanaPri(t *testing.T) {
	testK8sGrafana(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "pri")
}

func testK8sGrafana(t *testing.T, contextName string, envName string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)
	
	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix")) 
	namespaceName := fmt.Sprintf("grafana-%s", envName)
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	kubectlOptionsValues := terrak8s.NewKubectlOptions(contextName, "", "crossplane-system")
	CMValues := terrak8s.GetConfigMap(t, kubectlOptionsValues, "aws-crossplane")
	setValues["awsRegion"] = CMValues.Data["awsRegion"]
	setValues["awsAccount"] = CMValues.Data["awsAccount"]
	setValues["clusterOIDC"] = CMValues.Data["clusterOIDC"]

	if prefix != "runner-main" {
	   namespaceName = fmt.Sprintf("grafana-%s-%s", envName, prefix)
	   extraArgs["upgrade"] = []string{"--skip-crds"}
	   extraArgs["install"] = []string{"--skip-crds"}
	}
	releaseName := namespaceName

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
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, namespaceName, 10, 6*time.Second)
	if err != nil {
		t.Fatal("grafana deployment error:", err)
	}

}
