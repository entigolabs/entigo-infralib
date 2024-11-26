package test

import (

	"testing"

	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

const bucketName = "infralib-modules-aws-kms-tf"

var awsRegion string

func TestTerraformKms(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformKmsBiz)
	t.Run("Pri", testTerraformKmsPri)
}

func testTerraformKmsBiz(t *testing.T) {
	options := tf.InitAWSTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{})
	testTerraformKmsBizAssert(t, "biz", options)
}

func testTerraformKmsPri(t *testing.T) {
	options := tf.InitAWSTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_pri.tfvars", map[string]interface{}{})
	testTerraformKmsPriAssert(t, "pri", options)
}

func testTerraformKmsBizAssert(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc()

	assert.NotEmpty(t, outputs["telemetry_alias_arn"], "telemetry_alias_arn was not returned")
	assert.NotEmpty(t, outputs["config_alias_arn"], "config_alias_arn was not returned")
	assert.NotEmpty(t, outputs["data_alias_arn"], "data_alias_arn was not returned")

}

func testTerraformKmsPriAssert(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc()

	assert.Empty(t, outputs["telemetry_alias_arn"], "telemetry_alias_arn was not returned")
	assert.Empty(t, outputs["config_alias_arn"], "config_alias_arn was not returned")
	assert.Empty(t, outputs["data_alias_arn"], "data_alias_arn was not returned")

}
