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
	"github.com/stretchr/testify/require"
)

func TestK8sHelloWorldBiz(t *testing.T) {
	testK8sHelloWorld(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "./k8s_unit_basic_test_biz.yaml", "runner-main-biz-int.infralib.entigo.io")
}

func TestK8sHelloWorldPri(t *testing.T) {
	testK8sHelloWorld(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "./k8s_unit_basic_test_pri.yaml", "runner-main-pri.infralib.entigo.io")
}

func testK8sHelloWorld(t *testing.T, contextName string, valuesFile string, hostName string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := fmt.Sprintf("hello-world")
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	if prefix != "runner-main" {
		namespaceName = fmt.Sprintf("hello-world-%s", prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}
	releaseName := namespaceName

	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)

	helmOptions := &helm.Options{
		SetValues:         setValues,
		ValuesFiles:       []string{valuesFile},
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

	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s", namespaceName), 10, 6*time.Second)
	if err != nil {
		t.Fatal("hello-world deployment error:", err)
	}
}
