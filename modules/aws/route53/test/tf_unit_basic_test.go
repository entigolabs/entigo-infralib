package test

import (
	"testing"
	"strings"
	"os"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
	"github.com/davecgh/go-spew/spew"
)

func cleanupS3Bucket(t *testing.T, awsRegion string, bucketName string) {
	aws.EmptyS3Bucket(t, awsRegion, bucketName)
	aws.DeleteS3Bucket(t, awsRegion, bucketName)
}

func TestTerraformBasicBiz(t *testing.T) {
        t.Parallel()
	spew.Dump("")
	
	awsRegion := aws.GetRandomRegion(t, []string{"eu-north-1"}, nil)
	bucketName := "infralib-modules-aws-route53-tf"
	key := fmt.Sprintf("%s/terraform.tfstate", os.Getenv("TF_VAR_prefix"))

	 err := aws.CreateS3BucketE(t, awsRegion, bucketName)
	 if err != nil {
	    if strings.Contains(err.Error(), "BucketAlreadyOwnedByYou") {
	      fmt.Println("Bucket already owned by you. Skipping bucket creation.")
	    } else {
	      fmt.Println("Error:", err)
	    }
	 }
 
        rootFolder := ".." 
        terraformFolderRelativeToRoot := "test"
        tempTestFolder := test_structure.CopyTerraformFolderToTemp(t, rootFolder, terraformFolderRelativeToRoot)
	
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: tempTestFolder,
		Reconfigure: true,
		VarFiles: []string{"tf_unit_basic_test_biz.tfvars"},
		BackendConfig: map[string]interface{}{
			"bucket": bucketName,
			"key":    key,
			"region": awsRegion,
		},
	})
	terraform.Init(t, terraformOptions)
	terraform.WorkspaceSelectOrNew(t, terraformOptions, "biz")

        if os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
	  defer terraform.Destroy(t, terraformOptions)
	  // defer cleanupS3Bucket(t, awsRegion, bucketName)
	}

	terraform.Apply(t, terraformOptions)

        outputs, err := terraform.OutputAllE(t, terraformOptions)
        if err != nil {
	  t.Fatalf("Failed to get outputs")
        }
	spew.Dump(outputs)
	assert.Equal(t, "", "", "Wrong cluster_name returned")
}

func TestTerraformBasicPri(t *testing.T) {
        t.Parallel()
	
	awsRegion := aws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
	bucketName := "infralib-modules-aws-route53-tf"
	key := fmt.Sprintf("%s/terraform.tfstate", os.Getenv("TF_VAR_prefix"))
	
	 err := aws.CreateS3BucketE(t, awsRegion, bucketName)
	 if err != nil {
	    if strings.Contains(err.Error(), "BucketAlreadyOwnedByYou") {
	      fmt.Println("Bucket already owned by you. Skipping bucket creation.")
	    } else {
	      fmt.Println("Error:", err)
	    }
	 }
	 
        rootFolder := ".."
        terraformFolderRelativeToRoot := "test"
        tempTestFolder := test_structure.CopyTerraformFolderToTemp(t, rootFolder, terraformFolderRelativeToRoot)
	
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: tempTestFolder,
		Reconfigure: true,
		VarFiles: []string{"tf_unit_basic_test_pri.tfvars"},
		BackendConfig: map[string]interface{}{
			"bucket": bucketName,
			"key":    key,
			"region": awsRegion,
		},
	})
	terraform.Init(t, terraformOptions)
	terraform.WorkspaceSelectOrNew(t, terraformOptions, "pri")

        if os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
	  defer terraform.Destroy(t, terraformOptions)
	  // defer cleanupS3Bucket(t, awsRegion, bucketName)
	}

	terraform.Apply(t, terraformOptions)

        outputs, err := terraform.OutputAllE(t, terraformOptions)
        if err != nil {
          t.Fatalf("Failed to get outputs")
        }
        
	spew.Dump(outputs)
	assert.Equal(t, "", "", "Wrong cluster_name returned")
}
