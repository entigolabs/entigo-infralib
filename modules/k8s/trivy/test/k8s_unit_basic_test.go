package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sTrivyAWSPri(t *testing.T) {
	testK8sTrivy(t, "aws", "pri")
}

func testK8sTrivy(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, _ := k8s.CheckKubectlConnection(t, cloudName, envName)

	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("trivy-%s-trivy-operator", envName), 20, 6*time.Second)
	if err != nil {
		t.Fatal("trivy-operator deployment error:", err)
	}
}
