package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/helm"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

func TestK8sMimirBiz(t *testing.T) {
	testK8sMimir(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "biz", "runner-main-biz-int.infralib.entigo.io")
}

func TestK8sMimirPri(t *testing.T) {
	testK8sMimir(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "pri", "runner-main-pri.infralib.entigo.io")
}

func testK8sMimir(t *testing.T, contextName string, envName string, hostName string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := fmt.Sprintf("mimir-%s", envName)
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	awsRegion := aws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
	account := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/account", envName))
	clusteroidc := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/oidc_provider", envName))
	region := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/runner-main-%s/region", envName))

	setValues["global.region"] = region
	setValues["global.awsAccount"] = account
	setValues["global.clusterOIDC"] = clusteroidc

	if prefix != "runner-main" {
		namespaceName = fmt.Sprintf("mimir-%s-%s", envName, prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}
	releaseName := namespaceName
	bucketName := fmt.Sprintf("%s-logs", namespaceName)
	setValues["global.bucketName"] = bucketName

	setValues["mimir-distributed.gateway.ingress.hosts[0].host"] = fmt.Sprintf("%s.%s", releaseName, hostName)
	setValues["mimir-distributed.gateway.ingress.hosts[0].paths[0].path"] = "/"
	setValues["mimir-distributed.gateway.ingress.hosts[0].paths[0].pathType"] = "Prefix"

	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)

	helmOptions := &helm.Options{
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
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "mimir-gateway", 10, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-gateway deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "mimir-distributor", 10, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-distributor deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "mimir-querier", 10, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-querier deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "mimir-query-frontend", 10, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-query-frontend deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "mimir-ruler", 10, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-ruler deployment error:", err)
	}
	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, "mimir-compactor-0", 10, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-compactor-0 pod error:", err)
	}
	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, "mimir-ingester-0", 10, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-ingester-0 pod error:", err)
	}
	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, "mimir-store-gateway-0", 10, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-store-gateway-0 pod error:", err)
	}
}
