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

func TestK8sLokiBiz(t *testing.T) {
	testK8sLoki(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "biz")
}

func TestK8sLokiPri(t *testing.T) {
	testK8sLoki(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "pri")
}

func testK8sLoki(t *testing.T, contextName string, envName string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)
	
	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix")) 
	namespaceName := fmt.Sprintf("loki-%s", envName)
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)
	
	awsRegion := aws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
	account := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/account",envName))
	clusteroidc := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/oidc_provider",envName))
	region := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/region",envName))
	
	setValues["region"] = region
	
	
	setValues["awsAccount"] = account
	setValues["clusterOIDC"] = clusteroidc

	if prefix != "runner-main" {
	   namespaceName = fmt.Sprintf("loki-%s-%s", envName, prefix)
	   extraArgs["upgrade"] = []string{"--skip-crds"}
	   extraArgs["install"] = []string{"--skip-crds"}
	}
	releaseName := namespaceName
	bucketName := fmt.Sprintf("entigoinfralib-%s", namespaceName)
	setValues["loki.loki.storage.s3.region"] = region
	setValues["loki.loki.storage.bucketNames.chunks"] = releaseName
	setValues["loki.loki.storage.bucketNames.ruler"] = releaseName
	setValues["loki.loki.storage.bucketNames.admin"] = releaseName

	setValues["loki.loki.storage_config.aws.region"] = region
	setValues["loki.loki.storage_config.aws.bucketnames"] = releaseName
	setValues["loki.gateway.ingress.hosts[0].host"] = fmt.Sprintf("%s.runner-main-%s-int.infralib.entigo.io", releaseName, envName)
	setValues["loki.gateway.ingress.hosts[0].paths[0].path"] = "/"
	setValues["loki.gateway.ingress.hosts[0].paths[0].pathType"] = "Prefix"
	
	//setValues["promtail.config.clients[0].url"] = fmt.Sprintf("https://%s.runner-main-%s-int.infralib.entigo.io/loki/api/v1/push", releaseName, envName)

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
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "loki-gateway", 10, 6*time.Second)
	if err != nil {
		t.Fatal("loki-gateway deployment error:", err)
	}
	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, "loki-read-0", 10, 6*time.Second)
	if err != nil {
		t.Fatal("loki-read-0 pod error:", err)
	}
	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, "loki-write-0", 10, 6*time.Second)
	if err != nil {
		t.Fatal("loki-read-0 pod error:", err)
	}

}
