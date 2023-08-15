package test

import (
	"fmt"
	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

const bucketName = "infralib-modules-aws-vpc-tf"

var awsRegion string

func TestVPCRunner(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformBasicBiz)
	t.Run("Pri", testTerraformBasicPri)
}

func testTerraformBasicBiz(t *testing.T) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_biz.tfvars", "biz")
	defer destroyFunc()

	vpcId := outputs["vpc_id"]
	assert.NotEmpty(t, vpcId, "vpc_id was not returned")

	privateSubnets := fmt.Sprint(outputs["private_subnets"])
	assert.Equal(t, 3, len(strings.Split(privateSubnets, " ")), "Wrong number of private_subnets returned")

	publicSubnets := fmt.Sprint(outputs["public_subnets"])
	assert.Equal(t, 3, len(strings.Split(publicSubnets, " ")), "Wrong number of public_subnets returned")

	intraSubnets := fmt.Sprint(outputs["intra_subnets"])
	assert.Equal(t, "[]", intraSubnets, "Wrong number of intra_subnets returned")

	databaseSubnets := fmt.Sprint(outputs["database_subnets"])
	assert.Equal(t, 3, len(strings.Split(databaseSubnets, " ")), "Wrong number of database_subnets returned")

	databaseSubnetGroup := outputs["database_subnet_group"]
	assert.NotEmpty(t, databaseSubnetGroup, "database_subnet_group was not returned")

	elasticacheSubnets := fmt.Sprint(outputs["elasticache_subnets"])
	assert.Equal(t, 3, len(strings.Split(elasticacheSubnets, " ")), "Wrong number of elasticache_subnets returned")

	elasticacheSubnetGroup := outputs["elasticache_subnet_group"]
	assert.NotEmpty(t, elasticacheSubnetGroup, "elasticache_subnet_group was not returned")

	privateSubnetCidrs := fmt.Sprint(outputs["private_subnet_cidrs"])
	assert.Equal(t, "[10.146.32.0/21 10.146.40.0/21 10.146.48.0/21]", privateSubnetCidrs, "Wrong value for private_subnet_cidrs returned")

	publicSubnetCidrs := fmt.Sprint(outputs["public_subnet_cidrs"])
	assert.Equal(t, "[10.146.4.0/24 10.146.5.0/24 10.146.6.0/24]", publicSubnetCidrs, "Wrong value for public_subnet_cidrs returned")

	databaseSubnetCidrs := fmt.Sprint(outputs["database_subnet_cidrs"])
	assert.Equal(t, "[10.146.16.0/22 10.146.20.0/22 10.146.24.0/22]", databaseSubnetCidrs, "Wrong value for database_subnet_cidrs returned")

	elasticacheSubnetCidrs := fmt.Sprint(outputs["elasticache_subnet_cidrs"])
	assert.Equal(t, "[10.146.0.0/26 10.146.0.64/26 10.146.0.128/26]", elasticacheSubnetCidrs, "Wrong value for elasticache_subnet_cidrs returned")

	intraSubnetCidrs := fmt.Sprint(outputs["intra_subnet_cidrs"])
	assert.Equal(t, "[]", intraSubnetCidrs, "Wrong value for intra_subnet_cidrs returned")
}

func testTerraformBasicPri(t *testing.T) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_pri.tfvars", "pri")
	defer destroyFunc()

	vpcId := outputs["vpc_id"]
	assert.NotEmpty(t, vpcId, "vpc_id was not returned")

	privateSubnets := fmt.Sprint(outputs["private_subnets"])
	assert.Equal(t, 3, len(strings.Split(privateSubnets, " ")), "Wrong number of private_subnets returned")

	publicSubnets := fmt.Sprint(outputs["public_subnets"])
	assert.Equal(t, 2, len(strings.Split(publicSubnets, " ")), "Wrong number of public_subnets returned")

	intraSubnets := fmt.Sprint(outputs["intra_subnets"])
	assert.Equal(t, 1, len(strings.Split(intraSubnets, " ")), "Wrong number of intra_subnets returned")

	databaseSubnets := fmt.Sprint(outputs["database_subnets"])
	assert.Equal(t, "[]", databaseSubnets, "Wrong number of database_subnets returned")

	elasticacheSubnets := fmt.Sprint(outputs["elasticache_subnets"])
	assert.Equal(t, "[]", elasticacheSubnets, "Wrong number of elasticache_subnets returned")

	privateSubnetCidrs := fmt.Sprint(outputs["private_subnet_cidrs"])
	assert.Equal(t, "[10.146.32.0/21 10.146.40.0/21 10.146.48.0/21]", privateSubnetCidrs, "Wrong value for private_subnet_cidrs returned")

	publicSubnetCidrs := fmt.Sprint(outputs["public_subnet_cidrs"])
	assert.Equal(t, "[10.146.4.0/24 10.146.5.0/24]", publicSubnetCidrs, "Wrong value for public_subnet_cidrs returned")

	databaseSubnetCidrs := fmt.Sprint(outputs["database_subnet_cidrs"])
	assert.Equal(t, "[]", databaseSubnetCidrs, "Wrong value for database_subnet_cidrs returned")

	elasticacheSubnetCidrs := fmt.Sprint(outputs["elasticache_subnet_cidrs"])
	assert.Equal(t, "[]", elasticacheSubnetCidrs, "Wrong value for elasticache_subnet_cidrs returned")

	intraSubnetCidrs := fmt.Sprint(outputs["intra_subnet_cidrs"])
	assert.Equal(t, "[10.146.0.0/26]", intraSubnetCidrs, "Wrong value for intra_subnet_cidrs returned")
}
