package test

import (
	"testing"

	commonGoogle "github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

const bucketName = "infralib-modules-gce-vpc-tf"

var Region string

func TestTerraformVpc(t *testing.T) {
	Region = commonGoogle.SetupBucket(t, bucketName)
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

	assert.NotEmpty(t, outputs["database_subnets"], "Output database_subnets not returned")
	assert.NotEmpty(t, outputs["private_subnets"], "Output private_subnets not returned")
	assert.NotEmpty(t, outputs["nat_name"], "Output nat_name not returned")
	assert.NotEmpty(t, outputs["public_subnets"], "Output public_subnets not returned")
	assert.NotEmpty(t, outputs["router_id"], "Output router_id not returned")
	assert.NotEmpty(t, outputs["vpc_id"], "Output vpc_id not returned")

	assert.Len(t, outputs["database_subnets"], 2, "Wrong number of database_subnets returned")
	assert.Len(t, outputs["private_subnets"], 2, "Wrong number of private_subnets returned")
	assert.Len(t, outputs["intra_subnets"], 0, "Wrong number of intra_subnets returned")
	assert.Len(t, outputs["public_subnets"], 1, "Wrong number of public_subnets returned")

	assert.Equal(t, []interface{}{}, outputs["intra_subnet_cidrs"], "Wrong intra_subnet_cidrs returned")
	assert.Equal(t, []interface{}{"10.149.16.0/22", "10.149.20.0/22"}, outputs["database_subnet_cidrs"], "Wrong database_subnet_cidrs returned")
	assert.Equal(t, []interface{}{"10.149.32.0/22", "10.149.48.0/22"}, outputs["private_subnet_cidrs"], "Wrong private_subnet_cidrs returned")
	assert.Equal(t, []interface{}{"10.149.40.0/21", "10.149.56.0/21"}, outputs["private_subnet_cidrs_pods"], "Wrong private_subnet_cidrs_pods returned")
	assert.Equal(t, []interface{}{"10.149.36.0/22", "10.149.52.0/22"}, outputs["private_subnet_cidrs_services"], "Wrong private_subnet_cidrs_services returned")
	assert.Equal(t, []interface{}{"10.149.4.0/24"}, outputs["public_subnet_cidrs"], "Wrong public_subnet_cidrs returned")
}

func testTerraformVpcPriAssert(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc()

	assert.NotEmpty(t, outputs["database_subnets"], "Output database_subnets not returned")
	assert.NotEmpty(t, outputs["private_subnets"], "Output private_subnets not returned")
	assert.NotEmpty(t, outputs["nat_name"], "Output nat_name not returned")
	assert.NotEmpty(t, outputs["public_subnets"], "Output public_subnets not returned")
	assert.NotEmpty(t, outputs["router_id"], "Output router_id not returned")
	assert.NotEmpty(t, outputs["vpc_id"], "Output vpc_id not returned")

	assert.Len(t, outputs["database_subnets"], 1, "Wrong number of database_subnets returned")
	assert.Len(t, outputs["private_subnets"], 1, "Wrong number of private_subnets returned")
	assert.Len(t, outputs["intra_subnets"], 1, "Wrong number of intra_subnets returned")
	assert.Len(t, outputs["public_subnets"], 1, "Wrong number of public_subnets returned")

	assert.Equal(t, []interface{}{"10.29.40.0/22"}, outputs["intra_subnet_cidrs"], "Wrong intra_subnet_cidrs returned")
	assert.Equal(t, []interface{}{"10.29.48.0/21"}, outputs["database_subnet_cidrs"], "Wrong database_subnet_cidrs returned")
	assert.Equal(t, []interface{}{"10.29.0.0/21"}, outputs["private_subnet_cidrs"], "Wrong private_subnet_cidrs returned")
	assert.Equal(t, []interface{}{"10.29.16.0/20"}, outputs["private_subnet_cidrs_pods"], "Wrong private_subnet_cidrs_pods returned")
	assert.Equal(t, []interface{}{"10.29.8.0/21"}, outputs["private_subnet_cidrs_services"], "Wrong private_subnet_cidrs_services returned")
	assert.Equal(t, []interface{}{"10.29.32.0/21"}, outputs["public_subnet_cidrs"], "Wrong public_subnet_cidrs returned")
}
