package test

import (
	"testing"
	"strings"
	"os"
	"fmt"
	"time"
	"path/filepath"
	"github.com/gruntwork-io/terratest/modules/k8s"
        "github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	"github.com/davecgh/go-spew/spew"
)


func TestTerraformBasicBiz(t *testing.T) {
	spew.Dump("")
	
	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)
	
	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix")) 
	namespaceName := fmt.Sprintf("crossplane-system")
	releaseName := "crossplane"
	
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)
	
	if prefix != "runner-main" {
	   //releaseName = fmt.Sprintf("crossplane-%s", prefix)
	   extraArgs["upgrade"] = []string{"--skip-crds"}
	   extraArgs["install"] = []string{"--skip-crds"}
	}
	
	kubectlOptions := k8s.NewKubectlOptions("arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "", namespaceName)
	
	setValues["installProvider"] = "false"
	helmOptions := &helm.Options{
		SetValues: setValues,
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
	//https://entigo.atlassian.net/browse/RD-37
	//Add tests here that check if CRD is created
	time.Sleep(60 * time.Second)
	setValues["installProvider"] = "true"
	helmOptionsSecond := &helm.Options{
		SetValues: setValues,
		KubectlOptions:    kubectlOptions,
		BuildDependencies: false,
		ExtraArgs: extraArgs,
	}
	helm.Upgrade(t, helmOptionsSecond, helmChartPath, releaseName)
}


func TestTerraformBasicPri(t *testing.T) {
	spew.Dump("")
	
	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)
	
	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix")) 
	namespaceName := fmt.Sprintf("crossplane-system")
	releaseName := "crossplane"
	
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)
	
	if prefix != "runner-main" {
	   //releaseName = fmt.Sprintf("crossplane-%s", prefix)
	   extraArgs["upgrade"] = []string{"--skip-crds"}
	   extraArgs["install"] = []string{"--skip-crds"}
	}
	
	kubectlOptions := k8s.NewKubectlOptions("arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "", namespaceName)
	setValues["installProvider"] = "false"
	helmOptions := &helm.Options{
		SetValues: setValues,
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
	//https://entigo.atlassian.net/browse/RD-37
	//Add tests here that check if CRD is created
	time.Sleep(60 * time.Second)
	setValues["installProvider"] = "true"
	helmOptionsSecond := &helm.Options{
		SetValues: setValues,
		KubectlOptions:    kubectlOptions,
		BuildDependencies: false,
		ExtraArgs: extraArgs,
	}
	helm.Upgrade(t, helmOptionsSecond, helmChartPath, releaseName)


}
