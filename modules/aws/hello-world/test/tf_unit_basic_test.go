package test

import (
	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
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
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_biz.tfvars", "biz")
	defer destroyFunc()
	assert.Equal(t, outputs["hello_world"], fmt.Sprintf("Hello, %s-biz!", os.Getenv("TF_VAR_prefix")))
}

func testTerraformHelloWorldPri(t *testing.T) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_pri.tfvars", "pri")
	defer destroyFunc()
	assert.Equal(t, outputs["hello_world"], fmt.Sprintf("Hello, %s-pri!", os.Getenv("TF_VAR_prefix")))
}
