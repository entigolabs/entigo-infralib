package test

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/gruntwork-io/terratest/modules/helm"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const domain = "infralib.entigo.io"

func TestK8sAwsAlbBiz(t *testing.T) {
	testK8sAwsAlb(t, "aws-alb-biz", "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "runner-main-biz")
}

func TestK8sAwsAlbPri(t *testing.T) {
	testK8sAwsAlb(t, "aws-alb-pri", "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "runner-main-pri")
}

func testK8sAwsAlb(t *testing.T, namespaceName string, contextName string, runnerName string) {
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	kubectlOptionsValues := terrak8s.NewKubectlOptions(contextName, "", "crossplane-system")
	CMValues := terrak8s.GetConfigMap(t, kubectlOptionsValues, "aws-crossplane")
	setValues["aws-load-balancer-controller.image.repository"] = fmt.Sprintf("602401143452.dkr.ecr.%s.amazonaws.com/amazon/aws-load-balancer-controller", CMValues.Data["awsRegion"])
	setValues["awsAccount"] = CMValues.Data["awsAccount"]
	setValues["clusterOIDC"] = CMValues.Data["clusterOIDC"]
	setValues["aws-load-balancer-controller.clusterName"] = runnerName

	ingressClass := "alb"
	if prefix != "runner-main" {
		namespaceName = fmt.Sprintf("%s-%s", namespaceName, prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
		setValues["aws-load-balancer-controller.ingressClass"] = namespaceName
		ingressClass = namespaceName
	}
	setValues["aws-load-balancer-controller.nameOverride"] = namespaceName
	releaseName := namespaceName

	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)

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

	err = terrak8s.CreateNamespaceE(t, kubectlOptions, namespaceName)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println("Namespace already exists.")
		} else {
			t.Fatal("Error:", err)
		}
	}

	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)

	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, releaseName, 10, 6*time.Second)
	terrak8s.WaitUntilServiceAvailable(t, kubectlOptions, fmt.Sprintf("%s-webhook-service", releaseName), 60, 1*time.Second)
	time.Sleep(5 * time.Second)

	ingress, err := k8s.ReadObjectFromFile(t, "./templates/ingress.yaml")
	require.NoError(t, err)
	ingress.SetName(releaseName)
	err = unstructured.SetNestedField(ingress.Object, ingressClass, "spec", "ingressClassName")
	require.NoError(t, err, "Setting ingressClassName error")
	annotations := ingress.GetAnnotations()
	annotations["alb.ingress.kubernetes.io/group.name"] = releaseName
	ingress.SetAnnotations(annotations)
	host := fmt.Sprintf("%s.%s.%s", strings.ToLower(random.UniqueId()), runnerName, domain)
	err = k8s.SetNestedSliceString(ingress.Object, 0, "host", host, "spec", "rules")
	require.NoError(t, err, "Setting host error")
	createdIngress, err := k8s.CreateK8SIngress(t, kubectlOptions, ingress)
	require.NoError(t, err, "Creating ingress error")
	assert.NotNil(t, createdIngress, "Ingress is nil")
	assert.Equal(t, releaseName, createdIngress.GetName(), "Ingress name is not equal")

	_, err = k8s.WaitUntilK8SIngressAvailable(t, kubectlOptions, createdIngress.GetName(), 40, 2*time.Second)
	if err != nil {
		_ = k8s.DeleteK8SIngress(t, kubectlOptions, ingress.GetName()) // Try to delete ingress
	}
	require.NoError(t, err, "Ingress availability error")

	err = k8s.DeleteK8SIngress(t, kubectlOptions, ingress.GetName())
	require.NoError(t, err, "Deleting ingress error")

	err = k8s.WaitUntilK8SIngressDeleted(t, kubectlOptions, ingress.GetName(), 40, 2*time.Second)
	require.NoError(t, err, "Ingress didn't get deleted")
}
