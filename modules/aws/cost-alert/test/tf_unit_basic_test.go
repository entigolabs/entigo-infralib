package test

import (
	"testing"
        commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
        "github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

const bucketName = "infralib-modules-cost-alert-tf"

var awsRegion string

func TestCostAlertRunner(t *testing.T) {
        awsRegion = commonAWS.SetupBucket(t, bucketName)
        t.Run("Biz", testTerraformBasicBiz)
}

func testTerraformBasicBiz(t *testing.T) {
        t.Parallel()

        outputs, destroyFunc := tf.ApplyTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_biz.tfvars", "biz")
        defer destroyFunc()

	sns_topics := outputs["sns_topic_arns"]
	assert.NotEmpty(t, sns_topics, "sns_topic_arns was not returned")
}

