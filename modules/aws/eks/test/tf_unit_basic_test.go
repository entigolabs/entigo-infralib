package test

import (
	"fmt"
	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const bucketName = "infralib-modules-aws-eks-tf"

var awsRegion string

func TestTerraformEks(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformEksBiz)
	t.Run("Pri", testTerraformEksPri)
}

func testTerraformEksBiz(t *testing.T) {
	options := tf.InitTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_biz.tfvars")
	testTerraformEks(t, "biz", options)
}

func testTerraformEksPri(t *testing.T) {
	options := tf.InitTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_pri.tfvars")
	testTerraformEks(t, "pri", options)
}

func testTerraformEks(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc() // Defer needs to be called in outermost function
	clusterName := outputs["cluster_name"]
	assert.Equal(t, fmt.Sprintf("%s-%s", os.Getenv("TF_VAR_prefix"), workspaceName), clusterName,
		"Wrong cluster_name returned")
}
