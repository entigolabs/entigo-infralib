package test

import (
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/require"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sKeycloakAWSBiz(t *testing.T) {
	testK8sKeycloak(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks", "biz", "runner-main-biz-int.infralib.entigo.io", "aws")
}

func TestK8sKeycloakAWSPri(t *testing.T) {
	testK8sKeycloak(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks", "pri", "runner-main-pri.infralib.entigo.io", "aws")
}

func TestK8sKeycloakGoogleBiz(t *testing.T) {
	testK8sKeycloak(t, "gke_entigo-infralib2_europe-north1_biz-infra-gke", "biz", "runner-main-biz-int.gcp.infralib.entigo.io", "google")
}

func TestK8sKeycloakGooglePri(t *testing.T) {
	testK8sKeycloak(t, "gke_entigo-infralib2_europe-north1_pri-infra-gke", "pri", "runner-main-pri.gcp.infralib.entigo.io", "google")
}

func testK8sKeycloak(t *testing.T, contextName, envName, hostName, cloudProvider string) {
  	t.Parallel()
	namespaceName := fmt.Sprintf("keycloak-%s", envName)
        kubectlOptions := k8s.CheckKubectlConnection(t, contextName, namespaceName)


	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, fmt.Sprintf("%s-0", namespaceName), 30, 10*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s-0 pod error:", namespaceName), err)
	}
}
