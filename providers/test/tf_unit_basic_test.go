package test

import (
	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	testStructure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"testing"
	"fmt"
	"os"
)

const bucketName = "infralib-providers-tf"

var awsRegion string

func TestTerraformProviders(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	
	key := fmt.Sprintf("%s/terraform.tfstate", os.Getenv("TF_VAR_prefix"))

	tempTestFolder := testStructure.CopyTerraformFolderToTemp(t, "..", ".")
	err := os.Remove(fmt.Sprintf("%s/base.tf",tempTestFolder))
	if err != nil {
		fmt.Println("Error:", err)
	}

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: tempTestFolder,
		Reconfigure:  true,
		Vars: map[string]interface{}{
			"eks_cluster_name": "runner-main-biz",
			"gke_cluster_name": "runner-main-biz",
		},
		BackendConfig: map[string]interface{}{
			"bucket": bucketName,
			"key":    key,
			"region": awsRegion,
		},
	})
	defer terraform.Destroy(t, terraformOptions)
	terraform.Init(t, terraformOptions)
	terraform.Apply(t, terraformOptions)
}



