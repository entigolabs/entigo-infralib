package test

import (
	"testing"

	commonGCP "github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

const bucketName = "infralib-modules-gcp-dns-tf"

var Region string

func TestTerraformDns(t *testing.T) {
	Region = commonGCP.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformDnsBiz)
	t.Run("Pri", testTerraformDnsPri)
}

func testTerraformDnsBiz(t *testing.T) {
	options := tf.InitGCloudTerraform(t, bucketName, Region, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{})
	testTerraformDnsBizAssert(t, "biz", options)
}

func testTerraformDnsPri(t *testing.T) {
	options := tf.InitGCloudTerraform(t, bucketName, Region, "tf_unit_basic_test_Pri.tfvars", map[string]interface{}{})
	testTerraformDnsBizAssert(t, "biz", options)
}

func testTerraformDnsBizAssert(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc()
	_ = outputs
	assert.NotEmpty(t, outputs["google_project_iam_member"], "google_project_iam_member was not returned")
	assert.NotEmpty(t, outputs["dns_zone_name"], "dns_zone_name was not returned")
}

func testTerraformDnsPriAssert(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc()
	_ = outputs
	assert.NotEmpty(t, outputs["google_project_iam_member"], "google_project_iam_member was not returned")
	assert.NotEmpty(t, outputs["dns_zone_name"], "dns_zone_name was not returned")
}
