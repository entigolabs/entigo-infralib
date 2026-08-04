package test

import (
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

func TestIstioBaseAWSBiz(t *testing.T) {
	testIstioBase(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks")
}

func TestIstioBaseAWSPri(t *testing.T) {
	testIstioBase(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks")
}

func TestIstioBaseGoogleBiz(t *testing.T) {
	testIstioBase(t, "gke_entigo-infralib2_europe-north1_biz-infra-gke")
}

func TestIstioBaseGooglePri(t *testing.T) {
	testIstioBase(t, "gke_entigo-infralib2_europe-north1_pri-infra-gke")
}

func TestIstioBaseOracleDev(t *testing.T) {
	// Oracle's kubeconfig context isn't named after the cluster the way EKS/GKE ones are,
	// so it goes through the shared helper instead of a hardcoded context string.
	kubectlOptions, _ := k8s.CheckKubectlConnection(t, "oracle", "dev")
	kubectlOptions.Namespace = "istio-system"

	err := k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "networking.istio.io/v1beta1", []string{"virtualservices", "gateways"}, 60, 2*time.Second)
	require.NoError(t, err, "Istio Base CRDs missing")
}

func testIstioBase(t *testing.T, contextName string) {
	t.Parallel()
	namespaceName := "istio-system"
	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)

	err := k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "networking.istio.io/v1beta1", []string{"virtualservices"}, 60, 1*time.Second)
	require.NoError(t, err, "Istio Base no VirtualService CRD")
	err = k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "networking.istio.io/v1beta1", []string{"gateways"}, 60, 1*time.Second)
	require.NoError(t, err, "Istio Base no Gateway CRD")
}
