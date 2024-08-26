package test

import (
	"testing"

	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/terraform"
)

const bucketName = "infralib-modules-aws-eks-desired-size-tf"

var awsRegion string

func TestTerraformEksDesiredSize(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformEksDesiredSizeBiz)
	t.Run("Pri", testTerraformEksDesiredSizePri)
}

func testTerraformEksDesiredSizeBiz(t *testing.T) {
	options := tf.InitAWSTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{
		"cluster_name": "runner-main-biz",
	})
	testTerraformEksDesiredSize(t, "biz", options)
}

func testTerraformEksDesiredSizePri(t *testing.T) {
	options := tf.InitAWSTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_pri.tfvars", map[string]interface{}{
		"cluster_name": "runner-main-pri",
	})
	testTerraformEksDesiredSize(t, "pri", options)
}

func testTerraformEksDesiredSize(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	_, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc()
}
