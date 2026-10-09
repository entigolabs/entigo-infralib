package test

import (
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

func TestIstioIstiodAWSBiz(t *testing.T) {
	testIstioIstiod(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks")
}

func TestIstioIstiodAWSPri(t *testing.T) {
	testIstioIstiod(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks")
}

func TestIstioIstiodGoogleBiz(t *testing.T) {
	testIstioIstiod(t, "gke_entigo-infralib2_europe-north1_biz-infra-gke")
}

func TestIstioIstiodGooglePri(t *testing.T) {
	testIstioIstiod(t, "gke_entigo-infralib2_europe-north1_pri-infra-gke")
}

func TestIstioIstiodOracleBiz(t *testing.T) {
	testIstioIstiod(t, "biz-infra-oke")
}

func TestIstioIstiodOraclePri(t *testing.T) {
	testIstioIstiod(t, "pri-infra-oke")
}

func testIstioIstiod(t *testing.T, contextName string) {
	t.Parallel()
	namespaceName := "istio-system"

	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)
	output, err := terrak8s.RunKubectlAndGetOutputE(t, kubectlOptions, "auth", "can-i", "get", "pods")
	require.NoError(t, err, "Unable to connect to context %s cluster %s", contextName, err)
	require.Equal(t, output, "yes")

	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "istiod", 10, 6*time.Second)
	if err != nil {
		t.Fatal("istiod deployment error:", err)
	}
}
