package test

import (
	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
	"testing"
)

const bucketName = "infralib-modules-aws-helm-tf"

var awsRegion string

func TestHelmGit(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformHelmGitBiz)
}

func testTerraformHelmGitBiz(t *testing.T) {
	testTerraformHelmGit(t, "tf_unit_basic_test_biz.tfvars", "biz")
}

func testTerraformHelmGit(t *testing.T, varFile string, workspaceName string) {

	outputs, destroyFunc := tf.ApplyTerraform(t, bucketName, awsRegion, varFile, workspaceName)
	defer destroyFunc() // Defer needs to be called in outermost function
	status := outputs["status"]
	assert.Equal(t, "deployed", status,
		"Wrong status returned")
}
