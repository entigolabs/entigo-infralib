package test

import (
	"fmt"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/k8s"
)


const bucketName = "infralib-modules-aws-kms-tf"

func TestK8sAwsStorageclassBiz(t *testing.T) {
	testK8sAwsStorageclass(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks", "biz")
}

func TestK8sAwsStorageclassPri(t *testing.T) {
	testK8sAwsStorageclass(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks", "pri")
}

func testK8sAwsStorageclass(t *testing.T, contextName string, envName string) {
	t.Parallel()
	namespaceName := fmt.Sprintf("aws-storageclass-%s", envName)
        _ = k8s.CheckKubectlConnection(t, contextName, namespaceName)
	
	
}
