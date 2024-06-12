package test

import (
	"testing"

	commonGCP "github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, "10.0.0.0/16", outputs["subnet_cidr"], "Wrong cidr_block returned for subnet")
	assert.Equal(t, "10.4.0.0/14", outputs["subnet_cidr_pods"], "Wrong cidr_block returned for subnet_pods")
	assert.Equal(t, "10.8.0.0/20", outputs["subnet_cidr_services"], "Wrong cidr_block returned for subnet_services")
	assert.NotEmpty(t, outputs["nat_name"], "nat_name was not returned")
}

func testTerraformVpcPriAssert(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc()

	vpcId := outputs["vpc_id"]
	assert.NotEmpty(t, vpcId, "vpc_id was not returned")
}
