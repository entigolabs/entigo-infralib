package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sWireguardAWSBiz(t *testing.T) {
	testK8sWireguard(t, "aws", "biz")
}

func TestK8sWireguardAWSPri(t *testing.T) {
	testK8sWireguard(t, "aws", "pri")
}

func TestK8sWireguardGoogleBiz(t *testing.T) {
	testK8sWireguard(t, "google", "biz")
}

func TestK8sWireguardGooglePri(t *testing.T) {
	testK8sWireguard(t, "google", "pri")
}

func testK8sWireguard(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)

	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-wireguard", namespaceName), 30, 10*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("wireguard %s deployment error:", namespaceName), err)
	}
}
