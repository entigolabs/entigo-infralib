package test

import (
	"testing"
	"strings"
	"os"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/davecgh/go-spew/spew"
)


func TestTerraformBasicOne(t *testing.T) {
        // t.Parallel()
	
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "..",
		VarFiles: []string{"test/tf_unit_basic_test_1.tfvars"},
	})
	terraform.WorkspaceSelectOrNew(t, terraformOptions, "one")
        if os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
	  defer terraform.Destroy(t, terraformOptions)
	}

	terraform.InitAndApply(t, terraformOptions)
	vpc_id :=terraform.Output(t, terraformOptions, "vpc_id")
	assert.NotEmpty(t, vpc_id, "vpc_id was not returned")

	private_subnets :=terraform.Output(t, terraformOptions, "private_subnets")
	// spew.Dump(private_subnets)
	assert.Equal(t, 3, len(strings.Split(private_subnets, " ")), "Wrong number of private_subnets returned")
	  
	public_subnets :=terraform.Output(t, terraformOptions, "public_subnets")
	assert.Equal(t, 3, len(strings.Split(public_subnets, " ")), "Wrong number of public_subnets returned")
	  
	intra_subnets :=terraform.Output(t, terraformOptions, "intra_subnets")
	assert.Equal(t, "[]", intra_subnets, "Wrong number of intra_subnets returned")
	
	database_subnets :=terraform.Output(t, terraformOptions, "database_subnets")
	assert.Equal(t, 3, len(strings.Split(database_subnets, " ")), "Wrong number of database_subnets returned")
	
	database_subnet_group :=terraform.Output(t, terraformOptions, "database_subnet_group")
	assert.NotEmpty(t, database_subnet_group, "database_subnet_group was not returned")
	
	elasticache_subnets :=terraform.Output(t, terraformOptions, "elasticache_subnets")
	assert.Equal(t, 3, len(strings.Split(elasticache_subnets, " ")), "Wrong number of elasticache_subnets returned")
	
	elasticache_subnet_group :=terraform.Output(t, terraformOptions, "elasticache_subnet_group")
	assert.NotEmpty(t, elasticache_subnet_group, "elasticache_subnet_group was not returned")
	
	private_subnet_cidrs :=terraform.Output(t, terraformOptions, "private_subnet_cidrs")
	assert.Equal(t, "[10.146.32.0/21 10.146.40.0/21 10.146.48.0/21]", private_subnet_cidrs, "Wrong value for private_subnet_cidrs returned")
	  
	public_subnet_cidrs :=terraform.Output(t, terraformOptions, "public_subnet_cidrs")
	assert.Equal(t, "[10.146.4.0/24 10.146.5.0/24 10.146.6.0/24]", public_subnet_cidrs, "Wrong value for public_subnet_cidrs returned")
	
	database_subnet_cidrs :=terraform.Output(t, terraformOptions, "database_subnet_cidrs")
	assert.Equal(t, "[10.146.16.0/22 10.146.20.0/22 10.146.24.0/22]", database_subnet_cidrs, "Wrong value for database_subnet_cidrs returned")
	
	elasticache_subnet_cidrs :=terraform.Output(t, terraformOptions, "elasticache_subnet_cidrs")
	assert.Equal(t, "[10.146.0.0/26 10.146.0.64/26 10.146.0.128/26]", elasticache_subnet_cidrs, "Wrong value for elasticache_subnet_cidrs returned")
	
	intra_subnet_cidrs :=terraform.Output(t, terraformOptions, "intra_subnet_cidrs")
	assert.Equal(t, "[]", intra_subnet_cidrs, "Wrong value for intra_subnet_cidrs returned")

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

	vpc_id :=terraform.Output(t, terraformOptions, "vpc_id")
	assert.NotEmpty(t, vpc_id, "vpc_id was not returned")
	
	private_subnets :=terraform.Output(t, terraformOptions, "private_subnets")
	assert.Equal(t, 3, len(strings.Split(private_subnets, " ")), "Wrong number of private_subnets returned")
	  
	public_subnets :=terraform.Output(t, terraformOptions, "public_subnets")
	assert.Equal(t, 2, len(strings.Split(public_subnets, " ")), "Wrong number of public_subnets returned")
	  
	intra_subnets :=terraform.Output(t, terraformOptions, "intra_subnets")
	assert.Equal(t, 1, len(strings.Split(intra_subnets, " ")), "Wrong number of intra_subnets returned")
	
	database_subnets :=terraform.Output(t, terraformOptions, "database_subnets")
	assert.Equal(t, "[]", database_subnets, "Wrong number of database_subnets returned")
	
	elasticache_subnets :=terraform.Output(t, terraformOptions, "elasticache_subnets")
	assert.Equal(t, "[]", elasticache_subnets, "Wrong number of elasticache_subnets returned")
	
	private_subnet_cidrs :=terraform.Output(t, terraformOptions, "private_subnet_cidrs")
	assert.Equal(t, "[10.146.32.0/21 10.146.40.0/21 10.146.48.0/21]", private_subnet_cidrs, "Wrong value for private_subnet_cidrs returned")
	  
	public_subnet_cidrs :=terraform.Output(t, terraformOptions, "public_subnet_cidrs")
	assert.Equal(t, "[10.146.4.0/24 10.146.5.0/24]", public_subnet_cidrs, "Wrong value for public_subnet_cidrs returned")
	
	database_subnet_cidrs :=terraform.Output(t, terraformOptions, "database_subnet_cidrs")
	assert.Equal(t, "[]", database_subnet_cidrs, "Wrong value for database_subnet_cidrs returned")
	
	elasticache_subnet_cidrs :=terraform.Output(t, terraformOptions, "elasticache_subnet_cidrs")
	assert.Equal(t, "[]", elasticache_subnet_cidrs, "Wrong value for elasticache_subnet_cidrs returned")
	
	intra_subnet_cidrs :=terraform.Output(t, terraformOptions, "intra_subnet_cidrs")
	assert.Equal(t, "[10.146.0.0/26]", intra_subnet_cidrs, "Wrong value for intra_subnet_cidrs returned")
}
