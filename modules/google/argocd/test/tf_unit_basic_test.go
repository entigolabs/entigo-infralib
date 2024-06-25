package test

import (
	"testing"
        "os"
	"fmt"
	commonGCP "github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

const bucketName = "infralib-modules-gcp-argocd-tf"

var Region string
var prefix string

func TestTerraformArgocd(t *testing.T) {
	Region = commonGCP.SetupBucket(t, bucketName)
	prefix = os.Getenv("TF_VAR_prefix")
	t.Run("Biz", testTerraformArgocdBiz)
	t.Run("Pri", testTerraformArgocdPri)
}

func testTerraformArgocdBiz(t *testing.T) {
	namespace := "argocd-gcp"
	hostname := "argocd-gcp.runner-main-biz.gcp.infralib.entigo.io"
        if prefix != "runner-main" {
	  namespace = fmt.Sprintf("argocd-gcp-%s", prefix)
	  hostname = fmt.Sprintf("argocd-gcp-%s.runner-main-biz.gcp.infralib.entigo.io", prefix)
	}
	
	options := tf.InitGCloudTerraform(t, bucketName, Region, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{
		"hostname": hostname,
		"namespace": namespace,
		"name": namespace,
	})
	testTerraformArgocd(t, "biz", options)
}

func testTerraformArgocdPri(t *testing.T) {
	namespace := "argocd-gcp"
	hostname := "argocd-gcp.runner-main-pri.gcp.infralib.entigo.io"
        if prefix != "runner-main" {
	  namespace = fmt.Sprintf("argocd-gcp-%s", prefix)
	  hostname = fmt.Sprintf("argocd-gcp-%s.runner-main-pri.gcp.infralib.entigo.io", prefix)
	}
	
	options := tf.InitGCloudTerraform(t, bucketName, Region, "tf_unit_basic_test_pri.tfvars", map[string]interface{}{
		"hostname": hostname,
		"namespace": namespace,
		"name": namespace,
	})
	testTerraformArgocd(t, "pri", options)
}

func testTerraformArgocd(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc()

	assert.NotEmpty(t, outputs["hostname"], "name was not returned")
	assert.NotEmpty(t, outputs["namespace"], "name was not returned")
}
