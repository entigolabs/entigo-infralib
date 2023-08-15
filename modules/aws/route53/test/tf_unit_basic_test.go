package test

import (
	"github.com/davecgh/go-spew/spew"
	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"testing"
)

const bucketName = "infralib-modules-aws-route53-tf"

var awsRegion string

func TestRoute53Runner(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformBasicBiz)
	t.Run("Pri", testTerraformBasicPri)
}

func testTerraformBasicBiz(t *testing.T) {
	testTerraformBasic(t, "tf_unit_basic_test_biz.tfvars", "biz")
}

func testTerraformBasicPri(t *testing.T) {
	testTerraformBasic(t, "tf_unit_basic_test_pri.tfvars", "pri")
}

func testTerraformBasic(t *testing.T, varFile string, workspaceName string) {
	t.Parallel()
	outputs := tf.ApplyTerraform(t, bucketName, awsRegion, varFile, workspaceName)
	spew.Dump(outputs)
}
