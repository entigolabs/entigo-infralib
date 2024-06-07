package test

import (
	commonGCP "github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"testing"
)

const bucketName = "infralib-modules-gce-vpc-tf"

var Region string

func TestTerraformVpc(t *testing.T) {
	Region = commonGCP.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformVpcBiz)
	t.Run("Pri", testTerraformVpcPri)
}

func testTerraformVpcBiz(t *testing.T) {
	options := tf.InitGCloudTerraform(t, bucketName, Region, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{})
	testTerraformVpcBizAssert(t, "biz", options)
}

func testTerraformVpcPri(t *testing.T) {
	options := tf.InitGCloudTerraform(t, bucketName, Region, "tf_unit_basic_test_pri.tfvars", map[string]interface{}{})
	testTerraformVpcPriAssert(t, "pri", options)
}

func testTerraformVpcBizAssert(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc()

	vpcId := outputs["vpc_id"]
	assert.NotEmpty(t, vpcId, "vpc_id was not returned")

}

func testTerraformVpcPriAssert(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc()

	vpcId := outputs["vpc_id"]
	assert.NotEmpty(t, vpcId, "vpc_id was not returned")

}
