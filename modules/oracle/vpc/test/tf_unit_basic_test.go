package test

import (
	"testing"

	"github.com/entigolabs/entigo-infralib-common/oracle"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestTerraformVpc(t *testing.T) {
	t.Run("Biz", testTerraformVpcBiz)
}

func testTerraformVpcBiz(t *testing.T) {
	t.Parallel()
	outputs := oracle.GetTFOutputs(t, "biz")

	vpcId := tf.GetStringValue(t, outputs, "vpc__vpc_id")
	assert.NotEmpty(t, vpcId, "vpc_id was not returned")

	vpcCidr := tf.GetStringValue(t, outputs, "vpc__vpc_cidr")
	assert.Equal(t, "10.201.0.0/16", vpcCidr, "Wrong value for vpc_cidr returned")

	publicSubnets := tf.GetStringListValue(t, outputs, "vpc__public_subnets")
	assert.Equal(t, 1, len(publicSubnets), "Wrong number of public_subnets returned")

	privateSubnets := tf.GetStringListValue(t, outputs, "vpc__private_subnets")
	assert.Equal(t, 2, len(privateSubnets), "Wrong number of private_subnets returned")

	databaseSubnets := tf.GetStringListValue(t, outputs, "vpc__database_subnets")
	assert.Equal(t, 1, len(databaseSubnets), "Wrong number of database_subnets returned")

	intraSubnets := tf.GetStringListValue(t, outputs, "vpc__intra_subnets")
	assert.Equal(t, 1, len(intraSubnets), "Wrong number of intra_subnets returned")

	publicSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__public_subnet_cidrs")
	assert.Equal(t, "10.201.0.0/20", publicSubnetCidrs[0], "Wrong value for public_subnet_cidrs returned")

	privateSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__private_subnet_cidrs")
	assert.Equal(t, "10.201.16.0/20", privateSubnetCidrs[0], "Wrong value for private_subnet_cidrs returned")
	assert.Equal(t, "10.201.32.0/20", privateSubnetCidrs[1], "Wrong value for private_subnet_cidrs returned")

	databaseSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__database_subnet_cidrs")
	assert.Equal(t, "10.201.48.0/22", databaseSubnetCidrs[0], "Wrong value for database_subnet_cidrs returned")

	internetGatewayId := tf.GetStringValue(t, outputs, "vpc__internet_gateway_id")
	assert.NotEmpty(t, internetGatewayId, "Output internet_gateway_id not returned")

	natGatewayId := tf.GetStringValue(t, outputs, "vpc__nat_gateway_id")
	assert.NotEmpty(t, natGatewayId, "Output nat_gateway_id not returned")

	serviceGatewayId := tf.GetStringValue(t, outputs, "vpc__service_gateway_id")
	assert.NotEmpty(t, serviceGatewayId, "Output service_gateway_id not returned")
}
