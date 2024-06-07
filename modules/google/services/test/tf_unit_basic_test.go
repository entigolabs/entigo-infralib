package test

import (
	commonGCP "github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"testing"
)

const bucketName = "infralib-modules-gce-vpc-tf"

var Region string

func TestTerraformServices(t *testing.T) {
	Region = commonGCP.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformServicesBiz)
	t.Run("Pri", testTerraformServicesPri)
}

func testTerraformServicesBiz(t *testing.T) {
	options := tf.InitGCloudTerraform(t, bucketName, Region, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{})
	testTerraformServicesBizAssert(t, "biz", options)
}

func testTerraformServicesPri(t *testing.T) {
	options := tf.InitGCloudTerraform(t, bucketName, Region, "tf_unit_basic_test_pri.tfvars", map[string]interface{}{})
	testTerraformServicesPriAssert(t, "pri", options)
}

func testTerraformServicesBizAssert(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc()

}

func testTerraformServicesPriAssert(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc()

}
