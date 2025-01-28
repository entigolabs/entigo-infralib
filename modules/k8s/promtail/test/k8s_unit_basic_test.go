package test

import (
	"fmt"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/assert"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sPromtailAWSBiz(t *testing.T) {
	testK8sPromtail(t, "aws", "biz")
}

func TestK8sPromtailAWSPri(t *testing.T) {
	testK8sPromtail(t, "aws", "pri")
}

func TestK8sPromtailGoogleBiz(t *testing.T) {
	testK8sPromtail(t, "google", "biz")
}

func TestK8sPromtailGooglePri(t *testing.T) {
	testK8sPromtail(t, "google", "pri")
}

func testK8sPromtail(t *testing.T, cloudName string, envName string) {
  	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)

	dsname, err := terrak8s.GetDaemonSetE(t, kubectlOptions,  namespaceName)
	if err != nil {
		t.Fatal(fmt.Sprintf("daemonset %s error:", namespaceName), err)
	}
	assert.NotEmpty(t, dsname, "Daemonset was not returned")

}
