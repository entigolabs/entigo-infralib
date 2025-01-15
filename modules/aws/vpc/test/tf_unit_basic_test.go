package test

import (
	"testing"
	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/stretchr/testify/assert"
)

func TestTerraformVpc(t *testing.T) {
	t.Run("Biz", testTerraformVpcBiz)
	t.Run("Pri", testTerraformVpcPri)
}

func testTerraformVpcBiz(t *testing.T) {
        t.Parallel()
        outputs := aws.GetTFOutputs(t, "biz", "net")

	vpcId := aws.GetStringValue(t, outputs, "vpc__vpc_id")
	assert.NotEmpty(t, vpcId, "vpc_id was not returned")

	privateSubnets := aws.GetStringListValue(t, outputs, "vpc__private_subnets")
	assert.Equal(t, 2, len(privateSubnets), "Wrong number of private_subnets returned")

	publicSubnets := aws.GetStringListValue(t, outputs, "vpc__public_subnets")
	assert.Equal(t, 2, len(publicSubnets), "Wrong number of public_subnets returned")

	intraSubnets := aws.GetStringListValue(t, outputs, "vpc__intra_subnets")
	assert.Equal(t, 0, len(intraSubnets), "Wrong number of intra_subnets returned")

	databaseSubnets := aws.GetStringListValue(t, outputs, "vpc__database_subnets")
	assert.Equal(t, 2, len(databaseSubnets), "Wrong number of database_subnets returned")

	databaseSubnetGroup := aws.GetStringValue(t, outputs, "vpc__database_subnet_group")
	assert.NotEmpty(t, databaseSubnetGroup, "database_subnet_group was not returned")

	elasticacheSubnets := aws.GetStringListValue(t, outputs, "vpc__elasticache_subnets")
	assert.Equal(t, 2, len(elasticacheSubnets), "Wrong number of elasticache_subnets returned")

	elasticacheSubnetGroup := aws.GetStringValue(t, outputs, "vpc__elasticache_subnet_group")
	assert.NotEmpty(t, elasticacheSubnetGroup, "elasticache_subnet_group was not returned")

	privateSubnetCidrs := aws.GetStringListValue(t, outputs, "vpc__private_subnet_cidrs")
	assert.Equal(t, "10.146.32.0/21", privateSubnetCidrs[0], "Wrong value for private_subnet_cidrs returned")
	assert.Equal(t, "10.146.40.0/21", privateSubnetCidrs[1], "Wrong value for private_subnet_cidrs returned")

	publicSubnetCidrs := aws.GetStringListValue(t, outputs, "vpc__public_subnet_cidrs")
	assert.Equal(t, "10.146.4.0/24", publicSubnetCidrs[0], "Wrong value for public_subnet_cidrs returned")
	assert.Equal(t, "10.146.5.0/24", publicSubnetCidrs[1], "Wrong value for public_subnet_cidrs returned")

	databaseSubnetCidrs := aws.GetStringListValue(t, outputs, "vpc__database_subnet_cidrs")
	assert.Equal(t, "10.146.16.0/22", databaseSubnetCidrs[0], "Wrong value for database_subnet_cidrs returned")
	assert.Equal(t, "10.146.20.0/22", databaseSubnetCidrs[1], "Wrong value for database_subnet_cidrs returned")

	elasticacheSubnetCidrs := aws.GetStringListValue(t, outputs, "vpc__elasticache_subnet_cidrs")
	assert.Equal(t, "10.146.0.0/26", elasticacheSubnetCidrs[0], "Wrong value for elasticache_subnet_cidrs returned")
	assert.Equal(t, "10.146.0.64/26", elasticacheSubnetCidrs[1], "Wrong value for elasticache_subnet_cidrs returned")

	intraSubnetCidrs := aws.GetStringListValue(t, outputs, "vpc__intra_subnet_cidrs")
	assert.Equal(t, 0, len(intraSubnetCidrs), "Wrong value for intra_subnet_cidrs returned")
}

func testTerraformVpcPri(t *testing.T) {
	t.Parallel()
	outputs := aws.GetTFOutputs(t, "pri", "net")

	vpcId := aws.GetStringValue(t, outputs, "vpc__vpc_id")
	assert.NotEmpty(t, vpcId, "vpc_id was not returned")

	privateSubnets := aws.GetStringListValue(t, outputs, "vpc__private_subnets")
	assert.Equal(t, 3, len(privateSubnets), "Wrong number of private_subnets returned")

	publicSubnets := aws.GetStringListValue(t, outputs, "vpc__public_subnets")
	assert.Equal(t, 3, len(publicSubnets), "Wrong number of public_subnets returned")

	intraSubnets := aws.GetStringListValue(t, outputs, "vpc__intra_subnets")
	assert.Equal(t, 1, len(intraSubnets), "Wrong number of intra_subnets returned")

	databaseSubnets := aws.GetStringListValue(t, outputs, "vpc__database_subnets")
	assert.Equal(t, 3, len(databaseSubnets), "Wrong number of database_subnets returned")

	elasticacheSubnets := aws.GetStringListValue(t, outputs, "vpc__elasticache_subnets")
	assert.Equal(t, 0, len(elasticacheSubnets), "Wrong number of elasticache_subnets returned")

	privateSubnetCidrs := aws.GetStringListValue(t, outputs, "vpc__private_subnet_cidrs")
	assert.Equal(t, "10.24.16.0/22", privateSubnetCidrs[0], "Wrong value for private_subnet_cidrs returned")
	assert.Equal(t, "10.24.20.0/22", privateSubnetCidrs[1], "Wrong value for private_subnet_cidrs returned")
	assert.Equal(t, "10.24.24.0/22", privateSubnetCidrs[2], "Wrong value for private_subnet_cidrs returned")

	publicSubnetCidrs := aws.GetStringListValue(t, outputs, "vpc__public_subnet_cidrs")
	assert.Equal(t, "10.24.0.0/23", publicSubnetCidrs[0], "Wrong value for public_subnet_cidrs returned")
	assert.Equal(t, "10.24.2.0/23", publicSubnetCidrs[1], "Wrong value for public_subnet_cidrs returned")
	assert.Equal(t, "10.24.4.0/23", publicSubnetCidrs[2], "Wrong value for public_subnet_cidrs returned")

	databaseSubnetCidrs := aws.GetStringListValue(t, outputs, "vpc__database_subnet_cidrs")
	assert.Equal(t, "10.24.32.0/22", databaseSubnetCidrs[0], "Wrong value for database_subnet_cidrs returned")
	assert.Equal(t, "10.24.36.0/22", databaseSubnetCidrs[1], "Wrong value for database_subnet_cidrs returned")
	assert.Equal(t, "10.24.40.0/22", databaseSubnetCidrs[2], "Wrong value for database_subnet_cidrs returned")

	elasticacheSubnetCidrs := aws.GetStringListValue(t, outputs, "vpc__elasticache_subnet_cidrs")
	assert.Equal(t, 0, len(elasticacheSubnetCidrs), "Wrong number of elasticache_subnet_cidrs returned")

	intraSubnetCidrs := aws.GetStringListValue(t, outputs, "vpc__intra_subnet_cidrs")
	assert.Equal(t, "10.24.8.0/23", intraSubnetCidrs[0], "Wrong value for intra_subnet_cidrs returned")
}
