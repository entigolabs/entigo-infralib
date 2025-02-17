package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

const bucketName = "infralib-modules-aws-kms-tf"

func TestK8sAwsSamlProxyBiz(t *testing.T) {
	testK8sAwsSamlProxy(t, "aws", "biz")
}

func TestK8sAwsSamlProxyPri(t *testing.T) {
	testK8sAwsSamlProxy(t, "aws", "pri")
}

func testK8sAwsSamlProxy(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)
	
	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s", namespaceName), 20, 6*time.Second)
	if err != nil {
		t.Fatal("saml-proxy deployment error:", err)
	}
}

