package test

import (
	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
	"testing"
)

const bucketName = "infralib-modules-aws-crossplane-tf"

var awsRegion string

func TestTerraformCrossplane(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformCrossplaneBiz)
	t.Run("Pri", testTerraformCrossplanePri)
}

func testTerraformCrossplaneBiz(t *testing.T) {
        options := tf.InitTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{})
	testTerraformCrossplane(t, "biz", options)
}

func testTerraformCrossplanePri(t *testing.T) {
        options := tf.InitTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_pri.tfvars", map[string]interface{}{})
	testTerraformCrossplane(t, "pri", options)
}

func testTerraformCrossplane(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	assert.NotEqual(t, outputs, "", "outputs not defined")
	defer destroyFunc() // Defer needs to be called in outermost function
}
