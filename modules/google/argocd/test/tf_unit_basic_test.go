package test

import (
	"testing"

	commonGCP "github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

const bucketName = "infralib-modules-gcp-argocd-tf"

var Region string

func TestTerraformArgocd(t *testing.T) {
	Region = commonGCP.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformArgocdBiz)
	t.Run("Pri", testTerraformArgocdPri)
}

func testTerraformArgocdBiz(t *testing.T) {
	options := tf.InitGCloudTerraform(t, bucketName, Region, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{
		"gke_cluster_name": "runner-main-biz",
	})
	testTerraformArgocd(t, "biz", options)
}

func testTerraformArgocdPri(t *testing.T) {
	options := tf.InitGCloudTerraform(t, bucketName, Region, "tf_unit_basic_test_pri.tfvars", map[string]interface{}{
		"gke_cluster_name": "runner-main-pri",
	})
	testTerraformArgocd(t, "pri", options)
}

func testTerraformArgocd(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc()

	assert.NotEmpty(t, outputs["name"], "name was not returned")
}
