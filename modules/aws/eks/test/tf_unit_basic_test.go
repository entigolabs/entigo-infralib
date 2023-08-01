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


func TestTerraformBasicBiz(t *testing.T) {
        t.Parallel()
	spew.Dump("")
	awsRegion := aws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
	bucketName := "infralib-modules-aws-eks-tf"
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
	}
	terraform.Apply(t, terraformOptions)
	
	cluster_name := terraform.Output(t, terraformOptions, "cluster_name")
	assert.Equal(t, os.Getenv("TF_VAR_prefix") + "-one", cluster_name, "Wrong cluster_name returned")
}

func TestTerraformBasicPri(t *testing.T) {
        t.Parallel()
	awsRegion := aws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
	bucketName := "infralib-modules-aws-eks-tf"
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
	}
	terraform.Apply(t, terraformOptions)

	cluster_name := terraform.Output(t, terraformOptions, "cluster_name")
	assert.Equal(t, os.Getenv("TF_VAR_prefix") + "-two", cluster_name, "Wrong cluster_name returned")
}
