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

var (
	googleRegion  string
	vars          = make(map[string]interface{})
	prefix        = strings.ToLower(os.Getenv("TF_VAR_prefix"))
	googleProject = strings.ToLower(os.Getenv("GOOGLE_PROJECT"))
)

func TestTerraformCrossplane(t *testing.T) {
	googleRegion = commonGoogle.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformCrossplaneBiz)
	t.Run("Pri", testTerraformCrossplanePri)
}

func testTerraformCrossplaneBiz(t *testing.T) {
	envName := "biz"
	vars["crossplane_service_account_id"] = fmt.Sprintf("crossplane-%s", envName)

	if prefix != "runner-main" {
		vars["crossplane_service_account_id"] = fmt.Sprintf("crossplane-%s-%s", envName, prefix)
		vars["kubernetes_service_account"] = fmt.Sprintf("crossplane-%s-%s", prefix, envName)
		vars["kubernetes_namespace"] = fmt.Sprintf("crossplane-system-%s-%s", prefix, envName)
	}

	options := tf.InitGCloudTerraform(t, bucketName, googleRegion, fmt.Sprintf("tf_unit_basic_test_%s.tfvars", envName), vars)
	testTerraformCrossplane(t, envName, options)
}

func testTerraformCrossplanePri(t *testing.T) {
	envName := "pri"
	vars["crossplane_service_account_id"] = fmt.Sprintf("crossplane-%s", envName)

	if prefix != "runner-main" {
		vars["crossplane_service_account_id"] = fmt.Sprintf("crossplane-%s-%s", envName, prefix)
		vars["kubernetes_service_account"] = fmt.Sprintf("crossplane-%s-%s", prefix, envName)
		vars["kubernetes_namespace"] = fmt.Sprintf("crossplane-system-%s-%s", prefix, envName)
	}

	options := tf.InitGCloudTerraform(t, bucketName, googleRegion, fmt.Sprintf("tf_unit_basic_test_%s.tfvars", envName), vars)
	testTerraformCrossplane(t, envName, options)
}

func testTerraformCrossplane(t *testing.T, envName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, envName, options)

	googleServiceAccountId := truncateString(fmt.Sprintf("crossplane-%s", envName), 28)
	if prefix != "runner-main" {
		googleServiceAccountId = truncateString(fmt.Sprintf("crossplane-%s-%s", envName, prefix), 28)
	}

	assert.Equal(t, outputs["service_account_email"], fmt.Sprintf("%s@%s.iam.gserviceaccount.com", googleServiceAccountId, googleProject), "Wrong service_account_email returned")
	defer destroyFunc() // Defer needs to be called in outermost function
}

func truncateString(input string, maxLength int) string {
	if len(input) > maxLength {
		return input[:maxLength]
	}
	return input
}
