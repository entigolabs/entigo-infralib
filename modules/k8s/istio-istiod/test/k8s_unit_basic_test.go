package test

import (
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
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

func testIstioIstiod(t *testing.T, contextName string) {
  	t.Parallel()
	namespaceName := "istio-system"
        kubectlOptions := k8s.CheckKubectlConnection(t, contextName, namespaceName)


	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "istiod", 10, 6*time.Second)
	if err != nil {
		t.Fatal("istiod deployment error:", err)
	}
}
