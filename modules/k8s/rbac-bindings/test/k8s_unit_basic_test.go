package test

import (
	"testing"

)

const bucketName = "infralib-modules-aws-kms-tf"

func TestK8sAwsRbacBindingsBiz(t *testing.T) {
	testK8sAwsRbacBindings(t, "aws", "biz")
}

func TestK8sAwsRbacBindingsPri(t *testing.T) {
	testK8sAwsRbacBindings(t, "aws", "pri")
}

func testK8sAwsRbacBindings(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	
}

