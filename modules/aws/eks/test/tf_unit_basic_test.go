package test

import (
	"testing"
	"os"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/davecgh/go-spew/spew"
)


func TestTerraformBasicOne(t *testing.T) {
        // t.Parallel()
	spew.Dump("")
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "..",
		VarFiles: []string{"test/tf_unit_basic_test_1.tfvars"},
	})
	terraform.WorkspaceSelectOrNew(t, terraformOptions, "one")
        if os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
	  defer terraform.Destroy(t, terraformOptions)
	}

	terraform.InitAndApply(t, terraformOptions)
	cluster_name :=terraform.Output(t, terraformOptions, "cluster_name")
	assert.Equal(t, os.Getenv("TF_VAR_prefix") + "-one", cluster_name, "Wrong cluster_name returned")
	
	cluster_id :=terraform.Output(t, terraformOptions, "cluster_id")
	assert.NotEmpty(t, cluster_id, "cluster_id was not returned")


}

func TestTerraformBasicTwo(t *testing.T) {
        // t.Parallel()
	
	
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "..",
		VarFiles: []string{"test/tf_unit_basic_test_2.tfvars"},
	})
	terraform.WorkspaceSelectOrNew(t, terraformOptions, "two")
        if os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
	  defer terraform.Destroy(t, terraformOptions)
	}
	terraform.InitAndApply(t, terraformOptions)

	cluster_name :=terraform.Output(t, terraformOptions, "cluster_name")
	assert.Equal(t, os.Getenv("TF_VAR_prefix") + "-two", cluster_name, "Wrong cluster_name returned")
	
	cluster_id :=terraform.Output(t, terraformOptions, "cluster_id")
	assert.NotEmpty(t, cluster_id, "cluster_id was not returned")
}
