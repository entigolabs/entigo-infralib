package test

import (
	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
	"testing"
)

const bucketName = "infralib-modules-aws-argocd-tf"

var awsRegion string

func TestHelmGit(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformHelmGitBiz)
	t.Run("Pri", testTerraformHelmGitPri)
}

func testTerraformHelmGitBiz(t *testing.T) {
	testTerraformArgocdBootstrap(t, "tf_unit_basic_test_biz.tfvars", "biz")
}

func testTerraformHelmGitPri(t *testing.T) {
	testTerraformArgocdBootstrap(t, "tf_unit_basic_test_pri.tfvars", "pri")
}

func testTerraformArgocdBootstrap(t *testing.T, varFile string, workspaceName string) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, bucketName, awsRegion, varFile, workspaceName)
	defer destroyFunc() // Defer needs to be called in outermost function
	assert.NotEmpty(t, outputs["ssh-pub-key-id"],
		"No key fingerprint returned")
}
