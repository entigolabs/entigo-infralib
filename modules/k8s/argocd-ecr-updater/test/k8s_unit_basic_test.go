package test

import (
	"fmt"
	"testing"
	"time"
	//"os"
	//"strings"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sArgocdEcrUpdaterAWSBiz(t *testing.T) {
	testK8sArgocdEcrUpdater(t, "aws", "biz")
}

func TestK8sArgocdEcrUpdaterAWSPri(t *testing.T) {
	testK8sArgocdEcrUpdater(t, "aws", "pri")
}

func testK8sArgocdEcrUpdater(t *testing.T,  cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)
	


	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s", namespaceName), 20, 6*time.Second)
	if err != nil {
		t.Fatal("deployment error:", err)
	}
	


	
}
