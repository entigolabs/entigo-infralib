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

func TestK8sIstioGatewayBiz(t *testing.T) {
	testK8sIstioGateway(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz","./k8s_unit_basic_test_biz.yaml")
}

func TestK8sIstioGatewayPri(t *testing.T) {
	testK8sIstioGateway(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri","./k8s_unit_basic_test_pri.yaml")
}


func testK8sIstioGateway(t *testing.T, contextName string, valuesFile string) {
	spew.Dump("")
	
	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)
	
	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix")) 
	namespaceName := fmt.Sprintf("istio-gateway")
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)
	
	if prefix != "runner-main" {
	   namespaceName = fmt.Sprintf("istio-gateway-%s", prefix)
	   extraArgs["upgrade"] = []string{"--skip-crds"}
	   extraArgs["install"] = []string{"--skip-crds"}
	   
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
	err = k8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "istio-gateway", 10, 5*time.Second)
	if err != nil {
		t.Fatal("istio-gateway deployment error:", err)
	}

}
