package test

import (
	"fmt"
	"testing"
	"time"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sEntigoPortalAgentAWSBiz(t *testing.T) {
	testK8sEntigoPortalAgent(t, "aws", "biz")
}

func TestK8sEntigoPortalAgentAWSPri(t *testing.T) {
	testK8sEntigoPortalAgent(t, "aws", "pri")
}

//func TestK8sEntigoPortalAgentGoogleBiz(t *testing.T) {
//	testK8sEntigoPortalAgent(t, "gke_entigo-infralib2_europe-north1_biz-infra-gke", "biz")
//}

//func TestK8sEntigoPortalAgentGooglePri(t *testing.T) {
//	testK8sEntigoPortalAgent(t, "gke_entigo-infralib2_europe-north1_pri-infra-gke", "pri")
//}

func testK8sEntigoPortalAgent(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)
  

	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, namespaceName, 30, 10*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s deployment error:", namespaceName), err)
	}

}
