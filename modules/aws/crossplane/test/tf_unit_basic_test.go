package test

import (
	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
	"testing"
)

const bucketName = "infralib-modules-aws-eks-tf"

var awsRegion string

func TestTerraformCrossplane(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformCrossplaneBiz)
	t.Run("Pri", testTerraformCrossplanePri)
}

func testTerraformCrossplaneBiz(t *testing.T) {
	testTerraformCrossplane(t, "tf_unit_basic_test_biz.tfvars", "biz")
}

func testTerraformCrossplanePri(t *testing.T) {
	testTerraformCrossplane(t, "tf_unit_basic_test_pri.tfvars", "pri")
}

func testTerraformCrossplane(t *testing.T, varFile string, workspaceName string) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, bucketName, awsRegion, varFile, workspaceName)
	assert.NotEqual(t, outputs, "", "outputs not defined")
	defer destroyFunc() // Defer needs to be called in outermost function
}
