package test

import (
	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"testing"
)

const bucketName = "infralib-modules-aws-argocd-tf"

var awsRegion string

func TestTerraformArgocd(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformArgocdBiz)
	t.Run("Pri", testTerraformArgocdPri)
}

func testTerraformArgocdBiz(t *testing.T) {
	options := tf.InitAWSTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{
		"eks_cluster_name": "runner-main-biz",
	})
	testTerraformArgocd(t, "biz", options)
}

func testTerraformArgocdPri(t *testing.T) {
	options := tf.InitAWSTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_pri.tfvars", map[string]interface{}{
		"eks_cluster_name": "runner-main-pri",
	})
	testTerraformArgocd(t, "pri", options)
}

func testTerraformArgocd(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc() // Defer needs to be called in outermost function
	assert.NotEmpty(t, outputs["ssh-pub-key-id"],
		"No key fingerprint returned")
}
