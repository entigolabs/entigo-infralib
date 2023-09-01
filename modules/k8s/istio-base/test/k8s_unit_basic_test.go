package test

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/gruntwork-io/terratest/modules/helm"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIstioBaseBiz(t *testing.T) {
	testIstioBase(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz")
}

func TestIstioBasePri(t *testing.T) {
	testIstioBase(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri")
}

func testIstioBase(t *testing.T, contextName string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	namespaceName := fmt.Sprintf("istio-system")
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	releaseName := "istio-base"

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
	err = k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "networking.istio.io/v1beta1", []string{"virtualservices"}, 60, 1*time.Second)
	require.NoError(t, err, "Istio Base no VirtualService CRD")
	err = k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "networking.istio.io/v1beta1", []string{"gateways"}, 60, 1*time.Second)
	require.NoError(t, err, "Istio Base no Gateway CRD")

}
