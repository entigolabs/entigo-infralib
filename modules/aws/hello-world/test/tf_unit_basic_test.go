package test

import (
	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
	"testing"
	"fmt"
	"os"
)

const bucketName = "infralib-modules-aws-hello-world-tf"

var awsRegion string

func TestTerraformHelloWorld(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformHelloWorldBiz)
	t.Run("Pri", testTerraformHelloWorldPri)
}

func testTerraformHelloWorldBiz(t *testing.T) {
        options := tf.InitTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_biz.tfvars")
	testTerraformHelloWorldBizAssert(t, "biz", options)
}

func testTerraformHelloWorldPri(t *testing.T) {
        options := tf.InitTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_pri.tfvars")
	testTerraformHelloWorldPriAssert(t, "pri", options)
}

func testTerraformHelloWorldBizAssert(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc() // Defer needs to be called in outermost function
	assert.Equal(t, outputs["hello_world"], fmt.Sprintf("Hello, %s-biz!", os.Getenv("TF_VAR_prefix")))
}

func testTerraformHelloWorldPriAssert(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc() // Defer needs to be called in outermost function
	assert.Equal(t, outputs["hello_world"], fmt.Sprintf("Hello, %s-pri!", os.Getenv("TF_VAR_prefix")))
}
