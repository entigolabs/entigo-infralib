package test

import (
	"testing"

	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

const bucketName = "infralib-modules-cost-alert-root-us-tf-a"

var awsRegion string

func TestCostAlert(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformCostAlertBiz)
}

func testTerraformCostAlertBiz(t *testing.T) {
	options := tf.InitAWSTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{})
	testTerraformCostAlert(t, "biz", options)
}

func testTerraformCostAlert(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc()

	sns_topics := outputs["sns_topic_arns"]
	assert.NotEmpty(t, sns_topics, "sns_topic_arns was not returned")
}
