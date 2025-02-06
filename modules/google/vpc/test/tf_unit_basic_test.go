package test

import (
	"testing"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/entigolabs/entigo-infralib-common/google"
	"github.com/stretchr/testify/assert"
)


func TestTerraformVpc(t *testing.T) {
	t.Run("Biz", testTerraformVpcBiz)
	t.Run("Pri", testTerraformVpcPri)
}

func testTerraformVpcBiz(t *testing.T) {
        t.Parallel()
        outputs := google.GetTFOutputs(t, "biz")
	
	vpcId := tf.GetStringValue(t, outputs, "vpc__vpc_id")
	assert.NotEmpty(t, vpcId, "vpc_id was not returned")
	
	privateSubnets := tf.GetStringListValue(t, outputs, "vpc__private_subnets")
	assert.Equal(t, 2, len(privateSubnets), "Wrong number of private_subnets returned")

	publicSubnets := tf.GetStringListValue(t, outputs, "vpc__public_subnets")
	assert.Equal(t, 1, len(publicSubnets), "Wrong number of public_subnets returned")

	intraSubnets := tf.GetStringListValue(t, outputs, "vpc__intra_subnets")
	assert.Equal(t, 0, len(intraSubnets), "Wrong number of intra_subnets returned")

	databaseSubnets := tf.GetStringListValue(t, outputs, "vpc__database_subnets")
	assert.Equal(t, 2, len(databaseSubnets), "Wrong number of database_subnets returned")
	
	natName := tf.GetStringValue(t, outputs, "vpc__nat_name")
	assert.NotEmpty(t, natName, "Output nat_name not returned")
	
	routerId := tf.GetStringValue(t, outputs, "vpc__router_id")
	assert.NotEmpty(t, routerId, "Output router_id not returned")

	privateSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__private_subnet_cidrs")
	assert.Equal(t, "10.149.128.0/20", privateSubnetCidrs[0], "Wrong value for private_subnet_cidrs returned")
	assert.Equal(t, "10.149.192.0/20", privateSubnetCidrs[1], "Wrong value for private_subnet_cidrs returned")

	publicSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__public_subnet_cidrs")
	assert.Equal(t, "10.149.4.0/24", publicSubnetCidrs[0], "Wrong value for public_subnet_cidrs returned")

	databaseSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__database_subnet_cidrs")
	assert.Equal(t, "10.149.16.0/22", databaseSubnetCidrs[0], "Wrong value for database_subnet_cidrs returned")
	assert.Equal(t, "10.149.20.0/22", databaseSubnetCidrs[1], "Wrong value for database_subnet_cidrs returned")

	privateSubnetPodsCidrs := tf.GetStringListValue(t, outputs, "vpc__private_subnet_cidrs_pods")
	assert.Equal(t, "10.149.160.0/19", privateSubnetPodsCidrs[0], "Wrong value for private_subnet_cidrs_pods returned")
	assert.Equal(t, "10.149.224.0/19", privateSubnetPodsCidrs[1], "Wrong value for private_subnet_cidrs_pods returned")
	
	privateSubnetServicesCidrs := tf.GetStringListValue(t, outputs, "vpc__private_subnet_cidrs_services")
	assert.Equal(t, "10.149.144.0/20", privateSubnetServicesCidrs[0], "Wrong value for private_subnet_cidrs_services returned")
	assert.Equal(t, "10.149.208.0/20", privateSubnetServicesCidrs[1], "Wrong value for private_subnet_cidrs_services returned")

	intraSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__intra_subnet_cidrs")
	assert.Equal(t, 0, len(intraSubnetCidrs), "Wrong value for intra_subnet_cidrs returned")
}

func testTerraformVpcPri(t *testing.T) {
        t.Parallel()
	outputs := google.GetTFOutputs(t, "pri")
	
	vpcId := tf.GetStringValue(t, outputs, "vpc__vpc_id")
	assert.NotEmpty(t, vpcId, "vpc_id was not returned")
	
	privateSubnets := tf.GetStringListValue(t, outputs, "vpc__private_subnets")
	assert.Equal(t, 1, len(privateSubnets), "Wrong number of private_subnets returned")

	publicSubnets := tf.GetStringListValue(t, outputs, "vpc__public_subnets")
	assert.Equal(t, 1, len(publicSubnets), "Wrong number of public_subnets returned")

	intraSubnets := tf.GetStringListValue(t, outputs, "vpc__intra_subnets")
	assert.Equal(t, 1, len(intraSubnets), "Wrong number of intra_subnets returned")

	databaseSubnets := tf.GetStringListValue(t, outputs, "vpc__database_subnets")
	assert.Equal(t, 1, len(databaseSubnets), "Wrong number of database_subnets returned")
	
	natName := tf.GetStringValue(t, outputs, "vpc__nat_name")
	assert.NotEmpty(t, natName, "Output nat_name not returned")
	
	routerId := tf.GetStringValue(t, outputs, "vpc__router_id")
	assert.NotEmpty(t, routerId, "Output router_id not returned")

	privateSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__private_subnet_cidrs")
	assert.Equal(t, "10.29.0.0/21", privateSubnetCidrs[0], "Wrong value for private_subnet_cidrs returned")

	publicSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__public_subnet_cidrs")
	assert.Equal(t, "10.29.32.0/21", publicSubnetCidrs[0], "Wrong value for public_subnet_cidrs returned")

	databaseSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__database_subnet_cidrs")
	assert.Equal(t, "10.29.48.0/21", databaseSubnetCidrs[0], "Wrong value for database_subnet_cidrs returned")

	privateSubnetPodsCidrs := tf.GetStringListValue(t, outputs, "vpc__private_subnet_cidrs_pods")
	assert.Equal(t, "10.29.16.0/20", privateSubnetPodsCidrs[0], "Wrong value for private_subnet_cidrs_pods returned")
	
	privateSubnetServicesCidrs := tf.GetStringListValue(t, outputs, "vpc__private_subnet_cidrs_services")
	assert.Equal(t, "10.29.8.0/21", privateSubnetServicesCidrs[0], "Wrong value for private_subnet_cidrs_services returned")

	intraSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__intra_subnet_cidrs")
	assert.Equal(t, "10.29.40.0/22", intraSubnetCidrs[0], "Wrong value for intra_subnet_cidrs returned")
	
}
