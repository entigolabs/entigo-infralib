package test

import (
	"testing"
	"strings"
	"os"
	"fmt"
	"path/filepath"
	"github.com/gruntwork-io/terratest/modules/k8s"
        "github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	"github.com/davecgh/go-spew/spew"
	"time"
)

func TestK8sArgocdBiz(t *testing.T) {
	testK8sArgocd(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "./k8s_unit_basic_test_biz.yaml", "runner-main-biz-int.infralib.entigo.io")
}

func TestK8sArgocdPri(t *testing.T) {
	testK8sArgocd(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "./k8s_unit_basic_test_pri.yaml", "runner-main-pri.infralib.entigo.io")
}

func testK8sArgocd(t *testing.T, contextName string, valuesFile string, hostName string) {
	spew.Dump("")
	
	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)
	
	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix")) 
	namespaceName := fmt.Sprintf("argocd")
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)
	
	if prefix != "runner-main" {
	   namespaceName = fmt.Sprintf("argocd-%s", prefix)
	   extraArgs["upgrade"] = []string{"--skip-crds"}
	   extraArgs["install"] = []string{"--skip-crds"}
	   setValues["argocd.crds.install"] = "false"
 	   setValues["argocd.server.config.url"]=fmt.Sprintf("https://%s.%s", namespaceName,hostName)
  	   setValues["argocd.server.ingress.hosts[0]"]=fmt.Sprintf("%s.%s", namespaceName,hostName)
	   
	}
	releaseName := namespaceName
	
	kubectlOptions := k8s.NewKubectlOptions(contextName, "", namespaceName)
	
	helmOptions := &helm.Options{
		SetValues: setValues,
		ValuesFiles: []string{valuesFile},
		KubectlOptions:    kubectlOptions,
		BuildDependencies: false,
		ExtraArgs: extraArgs,
	}

        if os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
	    defer helm.Delete(t, helmOptions, releaseName, true)
	    //k8s.DeleteNamespace(t, kubectlOptions, namespaceName)
	}

	err = k8s.CreateNamespaceE(t, kubectlOptions, namespaceName)
	if err != nil {
	    if strings.Contains(err.Error(), "already exists") {
	      fmt.Println("Namespace already exists.")
	    } else {
	      t.Fatal("Error:", err)
	    }
	}	
	

	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)
	err = k8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "argocd", 10, 6*time.Second)
	if err != nil {
		t.Fatal("argocd deployment error:", err)
	}

}

