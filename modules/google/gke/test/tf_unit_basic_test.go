package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

const bucketName = "infralib-modules-gce-vpc-tf"

var Region string

func TestTerraformGke(t *testing.T) {
	Region = commonGCP.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformGkeBiz)
}

func testTerraformGkeBiz(t *testing.T) {
	// vpc_id := aws.GetParameter(t, awsRegion, "/entigo-infralib/runner-main-biz/vpc_id")
	// private_subnets := aws.GetParameter(t, awsRegion, "/entigo-infralib/runner-main-biz/private_subnets")
	// public_subnets := aws.GetParameter(t, awsRegion, "/entigo-infralib/runner-main-biz/public_subnets")
	// private_subnet_cidrs := aws.GetParameter(t, awsRegion, "/entigo-infralib/runner-main-biz/private_subnet_cidrs")

	// options := tf.InitAWSTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{
	// 	"vpc_id":               vpc_id,
	// 	"private_subnets":      fmt.Sprintf("[%s]", private_subnets),
	// 	"public_subnets":       fmt.Sprintf("[%s]", public_subnets),
	// 	"eks_api_access_cidrs": fmt.Sprintf("[%s]", private_subnet_cidrs),
	// })
	// testTerraformGke(t, "biz", options)
}

func testTerraformGke(t *testing.T, workspaceName string, options *terraform.Options) {
	// t.Parallel()
	// outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	// defer destroyFunc() // Defer needs to be called in outermost function
	// clusterName := outputs["cluster_name"]
	// assert.Equal(t, fmt.Sprintf("%s-%s", os.Getenv("TF_VAR_prefix"), workspaceName), clusterName,
	// 	"Wrong cluster_name returned")
}
