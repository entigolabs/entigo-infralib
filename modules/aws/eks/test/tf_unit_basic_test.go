package test

import (
	"testing"
	"strings"
	"os"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/davecgh/go-spew/spew"
)


func TestTerraformBasicOne(t *testing.T) {
        // t.Parallel()
	spew.Dump("")
	awsRegion := aws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
	bucketName := "infralib-modules-aws-vpc-tf"
	key := fmt.Sprintf("%s/terraform.tfstate", os.Getenv("TF_VAR_prefix"))
	
	err := aws.CreateS3BucketE(t, awsRegion, bucketName)
	if err != nil {
	    if strings.Contains(err.Error(), "BucketAlreadyOwnedByYou") {
	      fmt.Println("Bucket already owned by you. Skipping bucket creation.")
	    } else {
	      fmt.Println("Error:", err)
	    }
	}
	
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "..",
		VarFiles: []string{"test/tf_unit_basic_test_1.tfvars"},
		BackendConfig: map[string]interface{}{
			"bucket": bucketName,
			"key":    key,
			"region": awsRegion,
		},
	})
	terraform.Init(t, terraformOptions)
	terraform.WorkspaceSelectOrNew(t, terraformOptions, "one")

        if os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
	  defer terraform.Destroy(t, terraformOptions)
	}
	terraform.Apply(t, terraformOptions)
	
	cluster_name :=terraform.Output(t, terraformOptions, "cluster_name")
	assert.Equal(t, os.Getenv("TF_VAR_prefix") + "-one", cluster_name, "Wrong cluster_name returned")
	
	cluster_id :=terraform.Output(t, terraformOptions, "cluster_id")
	assert.NotEmpty(t, cluster_id, "cluster_id was not returned")


}

func TestTerraformBasicTwo(t *testing.T) {
        // t.Parallel()
	awsRegion := aws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
	bucketName := "infralib-modules-aws-vpc-tf"
	key := fmt.Sprintf("%s/terraform.tfstate", os.Getenv("TF_VAR_prefix"))
	
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "..",
		VarFiles: []string{"test/tf_unit_basic_test_2.tfvars"},
		BackendConfig: map[string]interface{}{
			"bucket": bucketName,
			"key":    key,
			"region": awsRegion,
		},
	})
	terraform.Init(t, terraformOptions)
	terraform.WorkspaceSelectOrNew(t, terraformOptions, "two")

        if os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
	  defer terraform.Destroy(t, terraformOptions)
	}
	terraform.Apply(t, terraformOptions)

	cluster_name :=terraform.Output(t, terraformOptions, "cluster_name")
	assert.Equal(t, os.Getenv("TF_VAR_prefix") + "-two", cluster_name, "Wrong cluster_name returned")
	
	cluster_id :=terraform.Output(t, terraformOptions, "cluster_id")
	assert.NotEmpty(t, cluster_id, "cluster_id was not returned")
}
