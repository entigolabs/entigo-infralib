package test

import (
	"fmt"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/assert"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sPromtailAWSBiz(t *testing.T) {
	testK8sPromtail(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks", "biz")
}

func TestK8sPromtailAWSPri(t *testing.T) {
	testK8sPromtail(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks", "pri")
}

func TestK8sPromtailGoogleBiz(t *testing.T) {
	testK8sPromtail(t, "gke_entigo-infralib2_europe-north1_biz-infra-gke", "biz")
}

func TestK8sPromtailGooglePri(t *testing.T) {
	testK8sPromtail(t, "gke_entigo-infralib2_europe-north1_pri-infra-gke", "pri")
}

func testK8sPromtail(t *testing.T, contextName string, envName string) {
  	t.Parallel()
	namespaceName := fmt.Sprintf("promtail-%s", envName)
        kubectlOptions := k8s.CheckKubectlConnection(t, contextName, namespaceName)

	dsname, err := terrak8s.GetDaemonSetE(t, kubectlOptions,  namespaceName)
	if err != nil {
		t.Fatal(fmt.Sprintf("daemonset %s error:", namespaceName), err)
	}
	assert.NotEmpty(t, dsname, "Daemonset was not returned")

}
