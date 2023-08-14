package test

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTerraformBasicBiz(t *testing.T) {
	testTerraformBasic(t, "aws-alb-biz", "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "runner-main-biz")
}

func TestTerraformBasicPri(t *testing.T) {
	testTerraformBasic(t, "aws-alb-pri", "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "runner-main-pri")
}

func testTerraformBasic(t *testing.T, namespaceName string, contextName string, runnerName string) {
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	kubectlOptionsValues := k8s.NewKubectlOptions(contextName, "", "crossplane-system")
	CMValues := k8s.GetConfigMap(t, kubectlOptionsValues, "aws-crossplane")
	setValues["aws-load-balancer-controller.image.repository"] = fmt.Sprintf("602401143452.dkr.ecr.%s.amazonaws.com/amazon/aws-load-balancer-controller", CMValues.Data["awsRegion"])
	setValues["awsAccount"] = CMValues.Data["awsAccount"]
	setValues["clusterOIDC"] = CMValues.Data["clusterOIDC"]
	setValues["aws-load-balancer-controller.clusterName"] = runnerName

	if prefix != "runner-main" {
		namespaceName = fmt.Sprintf("%s-%s", namespaceName, prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
		setValues["aws-load-balancer-controller.ingressClass"] = namespaceName
		setValues["aws-load-balancer-controller.nameOverride"] = namespaceName
	}
	releaseName := namespaceName

	kubectlOptions := k8s.NewKubectlOptions(contextName, "", namespaceName)

	helmOptions := &helm.Options{
		SetValues:         setValues,
		KubectlOptions:    kubectlOptions,
		BuildDependencies: false,
		ExtraArgs:         extraArgs,
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

	k8s.WaitUntilDeploymentAvailable(t, kubectlOptions, releaseName, 60, 1*time.Second)
}
