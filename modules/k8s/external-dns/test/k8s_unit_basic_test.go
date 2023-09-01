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


func TestK8sExternalDnsBiz(t *testing.T) {
	testK8sExternalDns(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz")
}

func TestK8sExternalDnsPri(t *testing.T) {
	testK8sExternalDns(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri")
}

func testK8sExternalDns(t *testing.T, contextName string) {
	spew.Dump("")
	
	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)
	
	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix")) 
	namespaceName := fmt.Sprintf("external-dns")
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)
	
	
	kubectlOptionsValues := k8s.NewKubectlOptions(contextName, "", "crossplane-system")
	CMValues := k8s.GetConfigMap(t, kubectlOptionsValues, "aws-crossplane")
	setValues["external-dns.env[0].value"] = CMValues.Data["awsRegion"]
	setValues["external-dns.env[0].name"] = "AWS_DEFAULT_REGION"
	setValues["awsAccount"] = CMValues.Data["awsAccount"]
	setValues["clusterOIDC"] = CMValues.Data["clusterOIDC"]
	
	
	if prefix != "runner-main" {
	   namespaceName = fmt.Sprintf("external-dns-%s", prefix)
	   extraArgs["upgrade"] = []string{"--skip-crds"}
	   extraArgs["install"] = []string{"--skip-crds"}
	}
	releaseName := namespaceName
	
	kubectlOptions := k8s.NewKubectlOptions(contextName, "", namespaceName)
	
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
	err = k8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, namespaceName, 10, 6*time.Second)
	if err != nil {
		t.Fatal("external-dns deployment error:", err)
	}

}
