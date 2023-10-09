package test

import (
	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
	"testing"
)

const bucketName = "infralib-modules-aws-helm-tf"

var awsRegion string

func TestHelmGit(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformHelmGitBiz)
	t.Run("Pri", testTerraformHelmGitPri)
}

func testTerraformHelmGitBiz(t *testing.T) {
        options := tf.InitTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{})
	testTerraformHelmGit(t, "biz", options)
}

func testTerraformHelmGitPri(t *testing.T) {
        options := tf.InitTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_pri.tfvars", map[string]interface{}{})
	testTerraformHelmGit(t, "pri", options)
}

func testTerraformHelmGit(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc() // Defer needs to be called in outermost function
	status := outputs["status"]
	assert.Equal(t, "deployed", status,
		"Wrong status returned")
}
