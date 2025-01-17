package test

import (
	"fmt"
	"os"
	"testing"

	commonGoogle "github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

const bucketName = "infralib-modules-gcp-dns-tf"

var Region string

func TestTerraformDns(t *testing.T) {
	Region = commonGoogle.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformDnsBiz)
	t.Run("Pri", testTerraformDnsPri)
}

func testTerraformDnsBiz(t *testing.T) {
	projectID := os.Getenv("GOOGLE_PROJECT")

	network := commonGoogle.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-biz-vpc_id/versions/latest", projectID))

	options := tf.InitGCloudTerraform(t, bucketName, Region, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{
		"vpc_ids": []string{network},
	})
	testTerraformDnsBizAssert(t, "biz", options)
}

func testTerraformDnsPri(t *testing.T) {
	options := tf.InitGCloudTerraform(t, bucketName, Region, "tf_unit_basic_test_pri.tfvars", map[string]interface{}{})
	testTerraformDnsBizAssert(t, "pri", options)
}

func testTerraformDnsBizAssert(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc()
	assert.NotEmpty(t, outputs["pub_zone_id"], "pub_zone_id was not returned")
}

func testTerraformDnsPriAssert(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc()
	assert.NotEmpty(t, outputs["pub_domain"], "pub_domain was not returned")
}
