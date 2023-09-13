package test

import (
	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
	"testing"
)

const bucketName = "infralib-modules-aws-argocd-tf"

var awsRegion string

func TestArgocd(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformArgocdBiz)
	t.Run("Pri", testTerraformArgocdPri)
}

func testTerraformArgocdBiz(t *testing.T) {
	testTerraformArgocd(t, "tf_unit_basic_test_biz.tfvars", "biz")
}

func testTerraformArgocdPri(t *testing.T) {
	testTerraformArgocd(t, "tf_unit_basic_test_pri.tfvars", "pri")
}

func testTerraformArgocd(t *testing.T, varFile string, workspaceName string) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, bucketName, awsRegion, varFile, workspaceName)
	defer destroyFunc() // Defer needs to be called in outermost function
	assert.NotEmpty(t, outputs["ssh-pub-key-id"],
		"No key fingerprint returned")
}
