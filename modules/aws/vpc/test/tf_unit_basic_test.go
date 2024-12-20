package test

import (
	"fmt"
	"strings"
	"testing"

	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

const bucketName = "infralib-modules-aws-vpc-tf"

var awsRegion string

func TestTerraformVpc(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformVpcBiz)
	t.Run("Pri", testTerraformVpcPri)
}

func testTerraformVpcBiz(t *testing.T) {
  
        telemetry_alias_arn := aws.GetParameter(t, awsRegion, "/entigo-infralib/runner-main-biz/telemetry_alias_arn")
  
	options := tf.InitAWSTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{
		"flow_log_cloudwatch_log_group_kms_key_id":      telemetry_alias_arn,
	})
	testTerraformVpcBizAssert(t, "biz", options)
}

func testTerraformVpcPri(t *testing.T) {
	options := tf.InitAWSTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_pri.tfvars", map[string]interface{}{})
	testTerraformVpcPriAssert(t, "pri", options)
}

func testTerraformVpcBizAssert(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc()

	vpcId := outputs["vpc_id"]
	assert.NotEmpty(t, vpcId, "vpc_id was not returned")

	privateSubnets := fmt.Sprint(outputs["private_subnets"])
	assert.Equal(t, 2, len(strings.Split(privateSubnets, " ")), "Wrong number of private_subnets returned")

	publicSubnets := fmt.Sprint(outputs["public_subnets"])
	assert.Equal(t, 2, len(strings.Split(publicSubnets, " ")), "Wrong number of public_subnets returned")

	intraSubnets := fmt.Sprint(outputs["intra_subnets"])
	assert.Equal(t, "[]", intraSubnets, "Wrong number of intra_subnets returned")

	databaseSubnets := fmt.Sprint(outputs["database_subnets"])
	assert.Equal(t, 2, len(strings.Split(databaseSubnets, " ")), "Wrong number of database_subnets returned")

	databaseSubnetGroup := outputs["database_subnet_group"]
	assert.NotEmpty(t, databaseSubnetGroup, "database_subnet_group was not returned")

	elasticacheSubnets := fmt.Sprint(outputs["elasticache_subnets"])
	assert.Equal(t, 2, len(strings.Split(elasticacheSubnets, " ")), "Wrong number of elasticache_subnets returned")

	elasticacheSubnetGroup := outputs["elasticache_subnet_group"]
	assert.NotEmpty(t, elasticacheSubnetGroup, "elasticache_subnet_group was not returned")

	privateSubnetCidrs := fmt.Sprint(outputs["private_subnet_cidrs"])
	assert.Equal(t, "[10.146.32.0/21 10.146.40.0/21]", privateSubnetCidrs, "Wrong value for private_subnet_cidrs returned")

	publicSubnetCidrs := fmt.Sprint(outputs["public_subnet_cidrs"])
	assert.Equal(t, "[10.146.4.0/24 10.146.5.0/24]", publicSubnetCidrs, "Wrong value for public_subnet_cidrs returned")

	databaseSubnetCidrs := fmt.Sprint(outputs["database_subnet_cidrs"])
	assert.Equal(t, "[10.146.16.0/22 10.146.20.0/22]", databaseSubnetCidrs, "Wrong value for database_subnet_cidrs returned")

	elasticacheSubnetCidrs := fmt.Sprint(outputs["elasticache_subnet_cidrs"])
	assert.Equal(t, "[10.146.0.0/26 10.146.0.64/26]", elasticacheSubnetCidrs, "Wrong value for elasticache_subnet_cidrs returned")

	intraSubnetCidrs := fmt.Sprint(outputs["intra_subnet_cidrs"])
	assert.Equal(t, "[]", intraSubnetCidrs, "Wrong value for intra_subnet_cidrs returned")
}

func testTerraformVpcPriAssert(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc()

	vpcId := outputs["vpc_id"]
	assert.NotEmpty(t, vpcId, "vpc_id was not returned")

	privateSubnets := fmt.Sprint(outputs["private_subnets"])
	assert.Equal(t, 3, len(strings.Split(privateSubnets, " ")), "Wrong number of private_subnets returned")

	publicSubnets := fmt.Sprint(outputs["public_subnets"])
	assert.Equal(t, 3, len(strings.Split(publicSubnets, " ")), "Wrong number of public_subnets returned")

	intraSubnets := fmt.Sprint(outputs["intra_subnets"])
	assert.Equal(t, 1, len(strings.Split(intraSubnets, " ")), "Wrong number of intra_subnets returned")

	databaseSubnets := fmt.Sprint(outputs["database_subnets"])
	assert.Equal(t, 3, len(strings.Split(databaseSubnets, " ")), "Wrong number of database_subnets returned")

	elasticacheSubnets := fmt.Sprint(outputs["elasticache_subnets"])
	assert.Equal(t, "[]", elasticacheSubnets, "Wrong number of elasticache_subnets returned")

	privateSubnetCidrs := fmt.Sprint(outputs["private_subnet_cidrs"])
	assert.Equal(t, "[10.24.16.0/22 10.24.20.0/22 10.24.24.0/22]", privateSubnetCidrs, "Wrong value for private_subnet_cidrs returned")

	publicSubnetCidrs := fmt.Sprint(outputs["public_subnet_cidrs"])
	assert.Equal(t, "[10.24.0.0/23 10.24.2.0/23 10.24.4.0/23]", publicSubnetCidrs, "Wrong value for public_subnet_cidrs returned")

	databaseSubnetCidrs := fmt.Sprint(outputs["database_subnet_cidrs"])
	assert.Equal(t, "[10.24.32.0/22 10.24.36.0/22 10.24.40.0/22]", databaseSubnetCidrs, "Wrong value for database_subnet_cidrs returned")

	elasticacheSubnetCidrs := fmt.Sprint(outputs["elasticache_subnet_cidrs"])
	assert.Equal(t, "[]", elasticacheSubnetCidrs, "Wrong value for elasticache_subnet_cidrs returned")

	intraSubnetCidrs := fmt.Sprint(outputs["intra_subnet_cidrs"])
	assert.Equal(t, "[10.24.8.0/23]", intraSubnetCidrs, "Wrong value for intra_subnet_cidrs returned")
}
