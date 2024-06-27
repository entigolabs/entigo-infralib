package test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	commonGoogle "github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

const bucketName = "infralib-modules-gcp-crossplane-tf"

var googleRegion string
var vars = make(map[string]interface{})

func TestTerraformCrossplane(t *testing.T) {
	googleRegion = commonGoogle.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformCrossplaneBiz)
	t.Run("Pri", testTerraformCrossplanePri)
}

func testTerraformCrossplaneBiz(t *testing.T) {
	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	if prefix != "runner-main" {
		vars["ksa_name"] = fmt.Sprintf("crossplane-%s-biz", prefix)
		vars["kns_name"] = fmt.Sprintf("crossplane-system-%s-biz", prefix)
	}
	options := tf.InitGCloudTerraform(t, bucketName, googleRegion, "tf_unit_basic_test_biz.tfvars", vars)
	testTerraformCrossplane(t, "biz", options)
}

func testTerraformCrossplanePri(t *testing.T) {
	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	if prefix != "runner-main" {
		vars["ksa_name"] = fmt.Sprintf("crossplane-%s-pri", prefix)
		vars["kns_name"] = fmt.Sprintf("crossplane-system-%s-pri", prefix)
	}
	options := tf.InitGCloudTerraform(t, bucketName, googleRegion, "tf_unit_basic_test_pri.tfvars", vars)
	testTerraformCrossplane(t, "pri", options)
}

func testTerraformCrossplane(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	assert.NotEmpty(t, outputs["service_account_email"], "service_account_email was not returned")
	defer destroyFunc() // Defer needs to be called in outermost function
}
