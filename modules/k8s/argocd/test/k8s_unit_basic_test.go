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
)


func TestTerraformBasicBiz(t *testing.T) {
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
	   
	}
	releaseName := namespaceName
	
	kubectlOptions := k8s.NewKubectlOptions("arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "", namespaceName)
	
	helmOptions := &helm.Options{
		SetValues: setValues,
		ValuesFiles: []string{"./k8s_unit_basic_test_biz.yaml"},
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

	

}

