package test

import (
	"fmt"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
)

// func TestK8sAlloyAWSBiz(t *testing.T) {
// 	testK8sAlloy(t, "aws", "biz")
// }

func TestK8sAlloyAWSPri(t *testing.T) {
	testK8sAlloy(t, "aws", "pri")
}

// func TestK8sAlloyGoogleBiz(t *testing.T) {
// 	testK8sAlloy(t, "google", "biz")
// }

func TestK8sAlloyGooglePri(t *testing.T) {
	testK8sAlloy(t, "google", "pri")
}

func testK8sAlloy(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)

	daemonSetName, err := terrak8s.GetDaemonSetE(t, kubectlOptions, namespaceName)
	if err != nil {
		t.Fatal(fmt.Sprintf("Daemonset %s error:", namespaceName), err)
	}
	assert.NotEmpty(t, daemonSetName, "Daemonset was not returned")
}
