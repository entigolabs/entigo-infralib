package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gruntwork-io/terratest/modules/helm"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/stretchr/testify/require"
)

func TestK8sClusterAutoscalerAWSBiz(t *testing.T) {
	testK8sClusterAutoscaler(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "biz", "aws", "runner-main-biz")
}

func TestK8sClusterAutoscalerAWSPri(t *testing.T) {
	testK8sClusterAutoscaler(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "pri", "aws", "runner-main-pri")
}


func testK8sClusterAutoscaler(t *testing.T, contextName, envName, cloudProvider string, runnerName string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := fmt.Sprintf("cluster-autoscaler-%s", envName)
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)
	

	awsRegion := aws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
	awsAccount := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/%s/account", runnerName))
	clusteroidc := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/%s/oidc_provider", runnerName))
	region := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/%s/region", runnerName))
	
	setValues["global.aws.account"] = awsAccount
	setValues["global.aws.clusterOIDC"] = clusteroidc
	setValues["cluster-autoscaler.awsRegion"] = region

        setValues["cluster-autoscaler.autoDiscovery.clusterName"]=runnerName
	if prefix != "runner-main" {
		namespaceName = fmt.Sprintf("%s-%s", namespaceName, prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}

	releaseName := namespaceName

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


	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-aws-cluster-autoscaler", namespaceName), 50, 6*time.Second)
	if err != nil {
		t.Fatal("aws-cluster-autoscaler deployment error:", err)
	}



}
