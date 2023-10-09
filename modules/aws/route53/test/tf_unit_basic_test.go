package test

import (
	"github.com/davecgh/go-spew/spew"
	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"testing"
)

const bucketName = "infralib-modules-aws-route53-tf"

var awsRegion string

func TestTerraformRoute53(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformRoute53Biz)
	t.Run("Pri", testTerraformRoute53Pri)
	t.Run("Min", testTerraformRoute53Min)
	t.Run("Ext", testTerraformRoute53Ext)
}

func testTerraformRoute53Biz(t *testing.T) {
	options := tf.InitTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{})
	testTerraformRoute53(t, "biz", options)
}

func testTerraformRoute53Pri(t *testing.T) {
	options := tf.InitTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_pri.tfvars", map[string]interface{}{})
	testTerraformRoute53(t, "pri", options)
}

func testTerraformRoute53Min(t *testing.T) {
	options := tf.InitTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_min.tfvars", map[string]interface{}{})
	testTerraformRoute53(t, "min", options)
}

func testTerraformRoute53Ext(t *testing.T) {
	options := tf.InitTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_ext.tfvars", map[string]interface{}{})  
	testTerraformRoute53(t, "ext", options)
}

func testTerraformRoute53(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc() // Defer needs to be called in outermost function
	spew.Dump(outputs)
}
