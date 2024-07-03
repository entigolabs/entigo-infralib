package test

import (
	"fmt"
	"strings"
	"testing"

	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

const bucketName = "infralib-modules-aws-ec2-tf"

var awsRegion string

func TestTerraformEc2(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformEc2Biz)
	t.Run("Pri", testTerraformEc2Pri)
}

func testTerraformEc2Biz(t *testing.T) {
	// vpc_id := aws.GetParameter(t, awsRegion, "/entigo-infralib/runner-main-biz/vpc_id")
	public_subnets := aws.GetParameter(t, awsRegion, "/entigo-infralib/runner-main-biz/public_subnets")
	zone_id := aws.GetParameter(t, awsRegion, "/entigo-infralib/runner-main-biz/int_zone_id")

	options := tf.InitAWSTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{
		"subnet_id":       fmt.Sprintf("%s", strings.Trim(strings.Split(public_subnets, ",")[0], "\"")),
		"route53_zone_id": zone_id,
	})
	testTerraformEc2(t, "biz", options)
}

func testTerraformEc2Pri(t *testing.T) {
	// vpc_id := aws.GetParameter(t, awsRegion, "/entigo-infralib/runner-main-biz/vpc_id")
	public_subnets := aws.GetParameter(t, awsRegion, "/entigo-infralib/runner-main-biz/public_subnets")
	zone_id := aws.GetParameter(t, awsRegion, "/entigo-infralib/runner-main-pri/pub_zone_id")

	options := tf.InitAWSTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_pri.tfvars", map[string]interface{}{
		"subnet_id":       fmt.Sprintf("%s", strings.Trim(strings.Split(public_subnets, ",")[0], "\"")),
		"route53_zone_id": zone_id,
	})
	testTerraformEc2(t, "pri", options)
}

func testTerraformEc2(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc() // Defer needs to be called in outermost function
	assert.NotEqual(t, outputs["private_dns"], "", "Wrong cluster_name returned")
}
