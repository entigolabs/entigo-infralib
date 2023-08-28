package test

import (
	"github.com/davecgh/go-spew/spew"
	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"testing"
)

const bucketName = "infralib-modules-aws-route53-tf"

var awsRegion string

func TestTerraformRoute53(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformRoute53Biz)
	t.Run("Pri", testTerraformRoute53Pri)
}

func testTerraformRoute53Biz(t *testing.T) {
	testTerraformRoute53(t, "tf_unit_basic_test_biz.tfvars", "biz")
}

func testTerraformRoute53Pri(t *testing.T) {
	testTerraformRoute53(t, "tf_unit_basic_test_pri.tfvars", "pri")
}

func testTerraformRoute53(t *testing.T, varFile string, workspaceName string) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, bucketName, awsRegion, varFile, workspaceName)
	defer destroyFunc() // Defer needs to be called in outermost function
	spew.Dump(outputs)
}
