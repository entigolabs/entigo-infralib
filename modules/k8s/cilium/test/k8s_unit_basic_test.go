package test

import (
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sCiliumOracleBiz(t *testing.T) {
	testK8sCilium(t, "oracle", "biz")
}

func TestK8sCiliumOraclePri(t *testing.T) {
	testK8sCilium(t, "oracle", "pri")
}

func testK8sCilium(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)

	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "cilium-operator", 30, 10*time.Second)
	if err != nil {
		t.Fatal("cilium-operator deployment error:", err)
	}
	_, err = terrak8s.GetDaemonSetE(t, kubectlOptions, "cilium")
	if err != nil {
		t.Fatalf("cilium daemonset in %s error: %s", namespaceName, err)
	}
}
