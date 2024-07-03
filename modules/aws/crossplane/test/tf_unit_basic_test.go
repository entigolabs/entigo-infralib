package test

import (
	"testing"

	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

const bucketName = "infralib-modules-aws-crossplane-tf"

var awsRegion string

func TestTerraformCrossplane(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformCrossplaneBiz)
	t.Run("Pri", testTerraformCrossplanePri)
}

func testTerraformCrossplaneBiz(t *testing.T) {
	oidc_provider := aws.GetParameter(t, awsRegion, "/entigo-infralib/runner-main-biz/oidc_provider")
	oidc_provider_arn := aws.GetParameter(t, awsRegion, "/entigo-infralib/runner-main-biz/oidc_provider_arn")
	options := tf.InitAWSTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{
		"eks_oidc_provider":     oidc_provider,
		"eks_oidc_provider_arn": oidc_provider_arn,
	})
	testTerraformCrossplane(t, "biz", options)
}

func testTerraformCrossplanePri(t *testing.T) {
	oidc_provider := aws.GetParameter(t, awsRegion, "/entigo-infralib/runner-main-pri/oidc_provider")
	oidc_provider_arn := aws.GetParameter(t, awsRegion, "/entigo-infralib/runner-main-pri/oidc_provider_arn")
	options := tf.InitAWSTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_pri.tfvars", map[string]interface{}{
		"eks_oidc_provider":     oidc_provider,
		"eks_oidc_provider_arn": oidc_provider_arn,
	})
	testTerraformCrossplane(t, "pri", options)
}

func testTerraformCrossplane(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	assert.NotEqual(t, outputs, "", "outputs not defined")
	defer destroyFunc() // Defer needs to be called in outermost function
}
