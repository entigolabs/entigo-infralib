package test

import (
	"testing"

	"github.com/entigolabs/entigo-infralib-common/k8s"
)


const bucketName = "infralib-modules-aws-kms-tf"

func TestK8sAwsStorageclassBiz(t *testing.T) {
	testK8sAwsStorageclass(t, "aws", "biz")
}

func TestK8sAwsStorageclassPri(t *testing.T) {
	testK8sAwsStorageclass(t, "aws", "pri")
}

func testK8sAwsStorageclass(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	
	_, _ = k8s.CheckKubectlConnection(t, cloudName, envName)
	
}
