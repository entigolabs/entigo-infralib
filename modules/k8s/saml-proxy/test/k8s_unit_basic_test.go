package test

import (
	"testing"

	"github.com/entigolabs/entigo-infralib-common/k8s"
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
	
	_, _ = k8s.CheckKubectlConnection(t, cloudName, envName)
	
}
