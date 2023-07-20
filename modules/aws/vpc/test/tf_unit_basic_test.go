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

func TestTerraformBasicOne(t *testing.T) {
        t.Parallel()
	spew.Dump("")
	
	awsRegion := aws.GetRandomRegion(t, []string{"eu-north-1"}, nil)
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
 
        rootFolder := ".." 
        terraformFolderRelativeToRoot := "test"
        tempTestFolder := test_structure.CopyTerraformFolderToTemp(t, rootFolder, terraformFolderRelativeToRoot)
	
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: tempTestFolder,
		Reconfigure: true,
		VarFiles: []string{"tf_unit_basic_test_1.tfvars"},
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
	  // defer cleanupS3Bucket(t, awsRegion, bucketName)
	}

	terraform.Apply(t, terraformOptions)

        outputs, err := terraform.OutputAllE(t, terraformOptions)
        if err != nil {
	  t.Fatalf("Failed to get outputs")
        }

	vpc_id := outputs["vpc_id"]
	assert.NotEmpty(t, vpc_id, "vpc_id was not returned")

	private_subnets := fmt.Sprint(outputs["private_subnets"])
	assert.Equal(t, 3, len(strings.Split(private_subnets, " ")), "Wrong number of private_subnets returned")
	  
	public_subnets := fmt.Sprint(outputs["public_subnets"])
	assert.Equal(t, 3, len(strings.Split(public_subnets, " ")), "Wrong number of public_subnets returned")
	  
	intra_subnets := fmt.Sprint(outputs["intra_subnets"])
	assert.Equal(t, "[]", intra_subnets, "Wrong number of intra_subnets returned")
	
	database_subnets := fmt.Sprint(outputs["database_subnets"])
	assert.Equal(t, 3, len(strings.Split(database_subnets, " ")), "Wrong number of database_subnets returned")
	
	database_subnet_group := outputs["database_subnet_group"]
	assert.NotEmpty(t, database_subnet_group, "database_subnet_group was not returned")
	
	elasticache_subnets := fmt.Sprint(outputs["elasticache_subnets"])
	assert.Equal(t, 3, len(strings.Split(elasticache_subnets, " ")), "Wrong number of elasticache_subnets returned")
	
	elasticache_subnet_group := outputs["elasticache_subnet_group"]
	assert.NotEmpty(t, elasticache_subnet_group, "elasticache_subnet_group was not returned")
	
	private_subnet_cidrs := fmt.Sprint(outputs["private_subnet_cidrs"])
	assert.Equal(t, "[10.146.32.0/21 10.146.40.0/21 10.146.48.0/21]", private_subnet_cidrs, "Wrong value for private_subnet_cidrs returned")
	  
	public_subnet_cidrs := fmt.Sprint(outputs["public_subnet_cidrs"])
	assert.Equal(t, "[10.146.4.0/24 10.146.5.0/24 10.146.6.0/24]", public_subnet_cidrs, "Wrong value for public_subnet_cidrs returned")
	
	database_subnet_cidrs := fmt.Sprint(outputs["database_subnet_cidrs"])
	assert.Equal(t, "[10.146.16.0/22 10.146.20.0/22 10.146.24.0/22]", database_subnet_cidrs, "Wrong value for database_subnet_cidrs returned")
	
	elasticache_subnet_cidrs := fmt.Sprint(outputs["elasticache_subnet_cidrs"])
	assert.Equal(t, "[10.146.0.0/26 10.146.0.64/26 10.146.0.128/26]", elasticache_subnet_cidrs, "Wrong value for elasticache_subnet_cidrs returned")
	
	intra_subnet_cidrs := fmt.Sprint(outputs["intra_subnet_cidrs"])
	assert.Equal(t, "[]", intra_subnet_cidrs, "Wrong value for intra_subnet_cidrs returned")

}

func TestTerraformBasicTwo(t *testing.T) {
        t.Parallel()
	
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
	 
        rootFolder := ".."
        terraformFolderRelativeToRoot := "test"
        tempTestFolder := test_structure.CopyTerraformFolderToTemp(t, rootFolder, terraformFolderRelativeToRoot)
	
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: tempTestFolder,
		Reconfigure: true,
		VarFiles: []string{"tf_unit_basic_test_2.tfvars"},
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
	  // defer cleanupS3Bucket(t, awsRegion, bucketName)
	}

	terraform.Apply(t, terraformOptions)

        outputs, err := terraform.OutputAllE(t, terraformOptions)
        if err != nil {
          t.Fatalf("Failed to get outputs")
        }

        vpc_id := outputs["vpc_id"]
	assert.NotEmpty(t, vpc_id, "vpc_id was not returned")
	
        private_subnets := fmt.Sprint(outputs["private_subnets"])
	assert.Equal(t, 3, len(strings.Split(private_subnets, " ")), "Wrong number of private_subnets returned")
	  
        public_subnets := fmt.Sprint(outputs["public_subnets"])
	assert.Equal(t, 2, len(strings.Split(public_subnets, " ")), "Wrong number of public_subnets returned")
	  
        intra_subnets := fmt.Sprint(outputs["intra_subnets"])
	assert.Equal(t, 1, len(strings.Split(intra_subnets, " ")), "Wrong number of intra_subnets returned")
	
        database_subnets := fmt.Sprint(outputs["database_subnets"])
	assert.Equal(t, "[]", database_subnets, "Wrong number of database_subnets returned")
	
        elasticache_subnets := fmt.Sprint(outputs["elasticache_subnets"])
	assert.Equal(t, "[]", elasticache_subnets, "Wrong number of elasticache_subnets returned")
	
        private_subnet_cidrs := fmt.Sprint(outputs["private_subnet_cidrs"])
	assert.Equal(t, "[10.146.32.0/21 10.146.40.0/21 10.146.48.0/21]", private_subnet_cidrs, "Wrong value for private_subnet_cidrs returned")
	  
        public_subnet_cidrs := fmt.Sprint(outputs["public_subnet_cidrs"])
	assert.Equal(t, "[10.146.4.0/24 10.146.5.0/24]", public_subnet_cidrs, "Wrong value for public_subnet_cidrs returned")
	
        database_subnet_cidrs := fmt.Sprint(outputs["database_subnet_cidrs"])
	assert.Equal(t, "[]", database_subnet_cidrs, "Wrong value for database_subnet_cidrs returned")
	
        elasticache_subnet_cidrs := fmt.Sprint(outputs["elasticache_subnet_cidrs"])
	assert.Equal(t, "[]", elasticache_subnet_cidrs, "Wrong value for elasticache_subnet_cidrs returned")
	
        intra_subnet_cidrs := fmt.Sprint(outputs["intra_subnet_cidrs"])
	assert.Equal(t, "[10.146.0.0/26]", intra_subnet_cidrs, "Wrong value for intra_subnet_cidrs returned")
}
