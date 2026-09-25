package test

import (
	"os"
	"testing"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/entigolabs/entigo-infralib-common/aws"
	awsSDK "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	terratestAws "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerraformVpc(t *testing.T) {
	t.Run("Biz", testTerraformVpcBiz)
	t.Run("Pri", testTerraformVpcPri)
	t.Run("Spoke", testTerraformVpcSpoke)
}

func testTerraformVpcBiz(t *testing.T) {
        t.Parallel()
        outputs := aws.GetTFOutputs(t, "biz")

	vpcId := tf.GetStringValue(t, outputs, "vpc__vpc_id")
	assert.NotEmpty(t, vpcId, "vpc_id was not returned")

	privateSubnets := tf.GetStringListValue(t, outputs, "vpc__private_subnets")
	assert.Equal(t, 2, len(privateSubnets), "Wrong number of private_subnets returned")

	publicSubnets := tf.GetStringListValue(t, outputs, "vpc__public_subnets")
	assert.Equal(t, 2, len(publicSubnets), "Wrong number of public_subnets returned")

	intraSubnets := tf.GetStringListValue(t, outputs, "vpc__intra_subnets")
	assert.Equal(t, 0, len(intraSubnets), "Wrong number of intra_subnets returned")

	databaseSubnets := tf.GetStringListValue(t, outputs, "vpc__database_subnets")
	assert.Equal(t, 2, len(databaseSubnets), "Wrong number of database_subnets returned")

	databaseSubnetGroup := tf.GetStringValue(t, outputs, "vpc__database_subnet_group")
	assert.NotEmpty(t, databaseSubnetGroup, "database_subnet_group was not returned")

	elasticacheSubnets := tf.GetStringListValue(t, outputs, "vpc__elasticache_subnets")
	assert.Equal(t, 2, len(elasticacheSubnets), "Wrong number of elasticache_subnets returned")

	elasticacheSubnetGroup := tf.GetStringValue(t, outputs, "vpc__elasticache_subnet_group")
	assert.NotEmpty(t, elasticacheSubnetGroup, "elasticache_subnet_group was not returned")

	privateSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__private_subnets_cidr_blocks")
	assert.Equal(t, "10.146.32.0/21", privateSubnetCidrs[0], "Wrong value for private_subnets_cidr_blocks returned")
	assert.Equal(t, "10.146.40.0/21", privateSubnetCidrs[1], "Wrong value for private_subnets_cidr_blocks returned")

	publicSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__public_subnets_cidr_blocks")
	assert.Equal(t, "10.146.4.0/24", publicSubnetCidrs[0], "Wrong value for public_subnets_cidr_blocks returned")
	assert.Equal(t, "10.146.5.0/24", publicSubnetCidrs[1], "Wrong value for public_subnets_cidr_blocks returned")

	databaseSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__database_subnets_cidr_blocks")
	assert.Equal(t, "10.146.16.0/22", databaseSubnetCidrs[0], "Wrong value for database_subnets_cidr_blocks returned")
	assert.Equal(t, "10.146.20.0/22", databaseSubnetCidrs[1], "Wrong value for database_subnets_cidr_blocks returned")

	elasticacheSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__elasticache_subnets_cidr_blocks")
	assert.Equal(t, "10.146.0.0/26", elasticacheSubnetCidrs[0], "Wrong value for elasticache_subnets_cidr_blocks returned")
	assert.Equal(t, "10.146.0.64/26", elasticacheSubnetCidrs[1], "Wrong value for elasticache_subnets_cidr_blocks returned")

	controlSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__control_subnets_cidr_blocks")
	assert.Equal(t, controlSubnetCidrs, privateSubnetCidrs, "vpc__control_subnets_cidr_blocks must be same as vpc__private_subnets_cidr_blocks")
	
	serviceSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__service_subnets_cidr_blocks")
	assert.Equal(t, serviceSubnetCidrs, privateSubnetCidrs, "vpc__service_subnets_cidr_blocks must be same as vpc__private_subnets_cidr_blocks")
	
	computeSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__compute_subnets_cidr_blocks")
	assert.Equal(t, computeSubnetCidrs, privateSubnetCidrs, "vpc__compute_subnets_cidr_blocks must be same as vpc__private_subnets_cidr_blocks")
	
	controlSubnets := tf.GetStringListValue(t, outputs, "vpc__control_subnets")
	assert.Equal(t, controlSubnets, privateSubnets, "vpc__control_subnets must be same as vpc__private_subnets")
	
	serviceSubnets := tf.GetStringListValue(t, outputs, "vpc__service_subnets")
	assert.Equal(t, serviceSubnets, privateSubnets, "vpc__service_subnets must be same as vpc__private_subnets")
	
	computeSubnets := tf.GetStringListValue(t, outputs, "vpc__compute_subnets")
	assert.Equal(t, computeSubnets, privateSubnets, "vpc__compute_subnets must be same as vpc__private_subnets")
	
	intraSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__intra_subnets_cidr_blocks")
	assert.Equal(t, 0, len(intraSubnetCidrs), "Wrong value for intra_subnets_cidr_blocks returned")

	vpcIpv6CidrBlock := tf.GetStringValue(t, outputs, "vpc__vpc_ipv6_cidr_block")
	assert.NotEmpty(t, vpcIpv6CidrBlock, "vpc_ipv6_cidr_block was not returned")

	publicSubnetsIpv6CidrBlocks := tf.GetStringListValue(t, outputs, "vpc__public_subnets_ipv6_cidr_blocks")
	assert.Equal(t, 2, len(publicSubnetsIpv6CidrBlocks), "Wrong number of public_subnets_ipv6_cidr_blocks returned")

	privateSubnetsIpv6CidrBlocks := tf.GetStringListValue(t, outputs, "vpc__private_subnets_ipv6_cidr_blocks")
	assert.Equal(t, 2, len(privateSubnetsIpv6CidrBlocks), "Wrong number of private_subnets_ipv6_cidr_blocks returned")

	databaseSubnetsIpv6CidrBlocks := tf.GetStringListValue(t, outputs, "vpc__database_subnets_ipv6_cidr_blocks")
	assert.Equal(t, 2, len(databaseSubnetsIpv6CidrBlocks), "Wrong number of database_subnets_ipv6_cidr_blocks returned")

	elasticacheSubnetsIpv6CidrBlocks := tf.GetStringListValue(t, outputs, "vpc__elasticache_subnets_ipv6_cidr_blocks")
	assert.Equal(t, 2, len(elasticacheSubnetsIpv6CidrBlocks), "Wrong number of elasticache_subnets_ipv6_cidr_blocks returned")

	privateIpv6EgressRouteIds := tf.GetStringListValue(t, outputs, "vpc__private_ipv6_egress_route_ids")
	assert.NotEmpty(t, privateIpv6EgressRouteIds, "private_ipv6_egress_route_ids was not returned")

	endpoints := getEndpoints(t, outputs)
	assert.ElementsMatch(t, []string{"s3", "ecr_api", "ecr_dkr", "ec2", "sts", "efs"}, keys(endpoints), "Wrong vpc_endpoints returned")
	assert.Contains(t, getEndpointPolicy(t, endpoints["efs"]), "CustomEfsPolicy", "Custom efs endpoint policy was not applied")
	for key, action := range map[string]string{"s3": "s3:*", "ecr_api": "ecr:*", "ecr_dkr": "ecr:*", "ec2": "ec2:*", "sts": "sts:*"} {
		policy := getEndpointPolicy(t, endpoints[key])
		assertServicePolicy(t, policy, key, action)
		assert.Contains(t, policy, "o-cipguslp2x", "%s endpoint policy must trust the organization", key)
	}
	assert.Contains(t, getEndpointPolicy(t, endpoints["s3"]), "starport-layer-bucket", "s3 endpoint policy must allow the ECR layer bucket")
	assert.Contains(t, getEndpointPolicy(t, endpoints["sts"]), "sts:AssumeRoleWithWebIdentity", "sts endpoint policy must allow IRSA")
}

func testTerraformVpcPri(t *testing.T) {
	t.Parallel()
	outputs := aws.GetTFOutputs(t, "pri")

	vpcId := tf.GetStringValue(t, outputs, "vpc__vpc_id")
	assert.NotEmpty(t, vpcId, "vpc_id was not returned")

	privateSubnets := tf.GetStringListValue(t, outputs, "vpc__private_subnets")
	assert.Equal(t, 3, len(privateSubnets), "Wrong number of private_subnets returned")

	publicSubnets := tf.GetStringListValue(t, outputs, "vpc__public_subnets")
	assert.Equal(t, 3, len(publicSubnets), "Wrong number of public_subnets returned")

	intraSubnets := tf.GetStringListValue(t, outputs, "vpc__intra_subnets")
	assert.Equal(t, 2, len(intraSubnets), "Wrong number of intra_subnets returned")

	databaseSubnets := tf.GetStringListValue(t, outputs, "vpc__database_subnets")
	assert.Equal(t, 2, len(databaseSubnets), "Wrong number of database_subnets returned")

	elasticacheSubnets := tf.GetStringListValue(t, outputs, "vpc__elasticache_subnets")
	assert.Equal(t, 0, len(elasticacheSubnets), "Wrong number of elasticache_subnets returned")

	privateSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__private_subnets_cidr_blocks")
	assert.Equal(t, "10.24.8.0/23", privateSubnetCidrs[0], "Wrong value for private_subnets_cidr_blocks returned")
	assert.Equal(t, "10.24.10.0/23", privateSubnetCidrs[1], "Wrong value for private_subnets_cidr_blocks returned")
	assert.Equal(t, "10.24.12.0/23", privateSubnetCidrs[2], "Wrong value for private_subnets_cidr_blocks returned")

	publicSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__public_subnets_cidr_blocks")
	assert.Equal(t, "10.24.0.0/24", publicSubnetCidrs[0], "Wrong value for public_subnets_cidr_blocks returned")
	assert.Equal(t, "10.24.1.0/24", publicSubnetCidrs[1], "Wrong value for public_subnets_cidr_blocks returned")
	assert.Equal(t, "10.24.2.0/24", publicSubnetCidrs[2], "Wrong value for public_subnets_cidr_blocks returned")

	databaseSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__database_subnets_cidr_blocks")
	assert.Equal(t, "10.24.14.0/24", databaseSubnetCidrs[0], "Wrong value for database_subnets_cidr_blocks returned")
	assert.Equal(t, "10.24.15.0/24", databaseSubnetCidrs[1], "Wrong value for database_subnets_cidr_blocks returned")
	
	controlSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__control_subnets_cidr_blocks")
	assert.Equal(t, controlSubnetCidrs, privateSubnetCidrs, "vpc__control_subnets_cidr_blocks must be same as vpc__private_subnets_cidr_blocks")
	
	serviceSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__service_subnets_cidr_blocks")
	assert.Equal(t, serviceSubnetCidrs, privateSubnetCidrs, "vpc__service_subnets_cidr_blocks must be same as vpc__private_subnets_cidr_blocks")
	
	computeSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__compute_subnets_cidr_blocks")
	assert.Equal(t, computeSubnetCidrs, privateSubnetCidrs, "vpc__compute_subnets_cidr_blocks must be same as vpc__private_subnets_cidr_blocks")

	controlSubnets := tf.GetStringListValue(t, outputs, "vpc__control_subnets")
	assert.Equal(t, controlSubnets, privateSubnets, "vpc__control_subnets must be same as vpc__private_subnets")
	
	serviceSubnets := tf.GetStringListValue(t, outputs, "vpc__service_subnets")
	assert.Equal(t, serviceSubnets, privateSubnets, "vpc__service_subnets must be same as vpc__private_subnets")
	
	computeSubnets := tf.GetStringListValue(t, outputs, "vpc__compute_subnets")
	assert.Equal(t, computeSubnets, privateSubnets, "vpc__compute_subnets must be same as vpc__private_subnets")

	elasticacheSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__elasticache_subnets_cidr_blocks")
	assert.Equal(t, 0, len(elasticacheSubnetCidrs), "Wrong number of elasticache_subnets_cidr_blocks returned")

	intraSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__intra_subnets_cidr_blocks")
	assert.Equal(t, "10.24.4.0/23", intraSubnetCidrs[0], "Wrong value for intra_subnets_cidr_blocks returned")
	assert.Equal(t, "10.24.6.0/23", intraSubnetCidrs[1], "Wrong value for intra_subnets_cidr_blocks returned")

	vpcIpv6CidrBlock := tf.GetStringValue(t, outputs, "vpc__vpc_ipv6_cidr_block")
	assert.Empty(t, vpcIpv6CidrBlock, "vpc_ipv6_cidr_block should be empty when ipv6 is disabled")

	publicSubnetsIpv6CidrBlocks := tf.GetStringListValue(t, outputs, "vpc__public_subnets_ipv6_cidr_blocks")
	assert.Equal(t, 0, len(publicSubnetsIpv6CidrBlocks), "public_subnets_ipv6_cidr_blocks should be empty when ipv6 is disabled")

	privateSubnetsIpv6CidrBlocks := tf.GetStringListValue(t, outputs, "vpc__private_subnets_ipv6_cidr_blocks")
	assert.Equal(t, 0, len(privateSubnetsIpv6CidrBlocks), "private_subnets_ipv6_cidr_blocks should be empty when ipv6 is disabled")

	endpoints := getEndpoints(t, outputs)
	assert.ElementsMatch(t, []string{"s3"}, keys(endpoints), "Wrong vpc_endpoints returned")
	s3Policy := getEndpointPolicy(t, endpoints["s3"])
	assertServicePolicy(t, s3Policy, "s3", "s3:*")
	assert.NotContains(t, s3Policy, "aws:PrincipalOrgID", "s3 endpoint policy must not have an organization statement without endpoint_policy_org_id")
}

func testTerraformVpcSpoke(t *testing.T) {
	t.Parallel()
	outputs := aws.GetTFOutputs(t, "spoke")

	privateSubnets := tf.GetStringListValue(t, outputs, "vpc__private_subnets")
	assert.Equal(t, 9, len(privateSubnets), "Wrong number of private_subnets returned")
	
	controlSubnets := tf.GetStringListValue(t, outputs, "vpc__control_subnets")
	assert.Equal(t, 3, len(controlSubnets), "Wrong number of control_subnets returned")
	
	serviceSubnets := tf.GetStringListValue(t, outputs, "vpc__service_subnets")
	assert.Equal(t, 3, len(serviceSubnets), "Wrong number of service_subnets returned")
	
	computeSubnets := tf.GetStringListValue(t, outputs, "vpc__compute_subnets")
	assert.Equal(t, 3, len(computeSubnets), "Wrong number of compute_subnets returned")

	assert.Equal(t,controlSubnets[0],  privateSubnets[0], "Control and private subnet ids not matching")
	assert.Equal(t,controlSubnets[1],  privateSubnets[1], "Control and private subnet ids not matching")
	assert.Equal(t,controlSubnets[2],  privateSubnets[2], "Control and private subnet ids not matching")
	assert.Equal(t,serviceSubnets[0],  privateSubnets[3], "Service and private subnet ids not matching")
	assert.Equal(t,serviceSubnets[1],  privateSubnets[4], "Service and private subnet ids not matching")
	assert.Equal(t,serviceSubnets[2],  privateSubnets[5], "Service and private subnet ids not matching")
	assert.Equal(t,computeSubnets[0],  privateSubnets[6], "Compute and private subnet ids not matching")
	assert.Equal(t,computeSubnets[1],  privateSubnets[7], "Compute and private subnet ids not matching")
	assert.Equal(t,computeSubnets[2],  privateSubnets[8], "Compute and private subnet ids not matching")
	
	publicSubnets := tf.GetStringListValue(t, outputs, "vpc__public_subnets")
	assert.Equal(t, 3, len(publicSubnets), "Wrong number of public_subnets returned")

	intraSubnets := tf.GetStringListValue(t, outputs, "vpc__intra_subnets")
	assert.Equal(t, 3, len(intraSubnets), "Wrong number of intra_subnets returned")

	databaseSubnets := tf.GetStringListValue(t, outputs, "vpc__database_subnets")
	assert.Equal(t, 3, len(databaseSubnets), "Wrong number of database_subnets returned")

	elasticacheSubnets := tf.GetStringListValue(t, outputs, "vpc__elasticache_subnets")
	assert.Equal(t, databaseSubnets, elasticacheSubnets, "spoke elasticache_subnets must reuse database_subnets")

	elasticacheSubnetGroup := tf.GetStringValue(t, outputs, "vpc__elasticache_subnet_group")
	assert.NotEmpty(t, elasticacheSubnetGroup, "spoke should expose an elasticache subnet group (reused database subnets)")

	privateSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__private_subnets_cidr_blocks")
	assert.Equal(t, "10.30.0.64/28", privateSubnetCidrs[0], "Wrong value for private_subnets_cidr_blocks returned")
	assert.Equal(t, "10.30.0.80/28", privateSubnetCidrs[1], "Wrong value for private_subnets_cidr_blocks returned")
	assert.Equal(t, "10.30.0.96/28", privateSubnetCidrs[2], "Wrong value for private_subnets_cidr_blocks returned")
	assert.Equal(t, "10.30.2.0/26", privateSubnetCidrs[3], "Wrong value for private_subnets_cidr_blocks returned")
	assert.Equal(t, "10.30.2.64/26", privateSubnetCidrs[4], "Wrong value for private_subnets_cidr_blocks returned")
	assert.Equal(t, "10.30.2.128/26", privateSubnetCidrs[5], "Wrong value for private_subnets_cidr_blocks returned")
	assert.Equal(t, "10.30.4.0/24", privateSubnetCidrs[6], "Wrong value for private_subnets_cidr_blocks returned")
	assert.Equal(t, "10.30.5.0/24", privateSubnetCidrs[7], "Wrong value for private_subnets_cidr_blocks returned")
	assert.Equal(t, "10.30.6.0/24", privateSubnetCidrs[8], "Wrong value for private_subnets_cidr_blocks returned")
	
	controlSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__control_subnets_cidr_blocks")
	assert.Equal(t, "10.30.0.64/28", controlSubnetCidrs[0], "Wrong value for control_subnets_cidr_blocks returned")
	assert.Equal(t, "10.30.0.80/28", controlSubnetCidrs[1], "Wrong value for control_subnets_cidr_blocks returned")
	assert.Equal(t, "10.30.0.96/28", controlSubnetCidrs[2], "Wrong value for control_subnets_cidr_blocks returned")
	
	serviceSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__service_subnets_cidr_blocks")
	assert.Equal(t, "10.30.2.0/26", serviceSubnetCidrs[0], "Wrong value for service_subnets_cidr_blocks returned")
	assert.Equal(t, "10.30.2.64/26", serviceSubnetCidrs[1], "Wrong value for service_subnets_cidr_blocks returned")
	assert.Equal(t, "10.30.2.128/26", serviceSubnetCidrs[2], "Wrong value for service_subnets_cidr_blocks returned")
	
	computeSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__compute_subnets_cidr_blocks")
	assert.Equal(t, "10.30.4.0/24", computeSubnetCidrs[0], "Wrong value for compute_subnets_cidr_blocks returned")
	assert.Equal(t, "10.30.5.0/24", computeSubnetCidrs[1], "Wrong value for compute_subnets_cidr_blocks returned")
	assert.Equal(t, "10.30.6.0/24", computeSubnetCidrs[2], "Wrong value for compute_subnets_cidr_blocks returned")

	publicSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__public_subnets_cidr_blocks")
	assert.Equal(t, "10.30.1.0/26", publicSubnetCidrs[0], "Wrong value for public_subnets_cidr_blocks returned")
	assert.Equal(t, "10.30.1.64/26", publicSubnetCidrs[1], "Wrong value for public_subnets_cidr_blocks returned")
	assert.Equal(t, "10.30.1.128/26", publicSubnetCidrs[2], "Wrong value for public_subnets_cidr_blocks returned")

	databaseSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__database_subnets_cidr_blocks")
	assert.Equal(t, "10.30.3.0/26", databaseSubnetCidrs[0], "Wrong value for database_subnets_cidr_blocks returned")
	assert.Equal(t, "10.30.3.64/26", databaseSubnetCidrs[1], "Wrong value for database_subnets_cidr_blocks returned")
	assert.Equal(t, "10.30.3.128/26", databaseSubnetCidrs[2], "Wrong value for database_subnets_cidr_blocks returned")

	elasticacheSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__elasticache_subnets_cidr_blocks")
	assert.Equal(t, databaseSubnetCidrs, elasticacheSubnetCidrs, "spoke elasticache_subnets_cidr_blocks must reuse database_subnets_cidr_blocks")

	intraSubnetCidrs := tf.GetStringListValue(t, outputs, "vpc__intra_subnets_cidr_blocks")
	assert.Equal(t, "10.30.0.0/28", intraSubnetCidrs[0], "Wrong value for intra_subnets_cidr_blocks returned")
	assert.Equal(t, "10.30.0.16/28", intraSubnetCidrs[1], "Wrong value for intra_subnets_cidr_blocks returned")
	assert.Equal(t, "10.30.0.32/28", intraSubnetCidrs[2], "Wrong value for intra_subnets_cidr_blocks returned")

	vpcIpv6CidrBlock := tf.GetStringValue(t, outputs, "vpc__vpc_ipv6_cidr_block")
	assert.NotEmpty(t, vpcIpv6CidrBlock, "vpc_ipv6_cidr_block was not returned")

	publicSubnetsIpv6CidrBlocks := tf.GetStringListValue(t, outputs, "vpc__public_subnets_ipv6_cidr_blocks")
	assert.Equal(t, 3, len(publicSubnetsIpv6CidrBlocks), "Wrong number of public_subnets_ipv6_cidr_blocks returned")

	privateSubnetsIpv6CidrBlocks := tf.GetStringListValue(t, outputs, "vpc__private_subnets_ipv6_cidr_blocks")
	assert.Equal(t, 9, len(privateSubnetsIpv6CidrBlocks), "Wrong number of private_subnets_ipv6_cidr_blocks returned")

	privateIpv6EgressRouteIds := tf.GetStringListValue(t, outputs, "vpc__private_ipv6_egress_route_ids")
	assert.NotEmpty(t, privateIpv6EgressRouteIds, "private_ipv6_egress_route_ids was not returned")

	endpoints := getEndpoints(t, outputs)
	assert.ElementsMatch(t, []string{"sts"}, keys(endpoints), "Wrong vpc_endpoints returned, sts only must still create the endpoints module")
	assertServicePolicy(t, getEndpointPolicy(t, endpoints["sts"]), "sts", "sts:*")
}

func getEndpoints(t *testing.T, outputs map[string]interface{}) map[string]string {
	value, ok := tf.GetValue(t, outputs, "vpc__vpc_endpoints").(map[string]interface{})
	require.True(t, ok, "vpc__vpc_endpoints is not a map")
	result := make(map[string]string, len(value))
	for k, v := range value {
		result[k] = v.(string)
	}
	return result
}

func keys(m map[string]string) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

func getEndpointPolicy(t *testing.T, endpointId string) string {
	client := terratestAws.NewEc2Client(t, os.Getenv("AWS_REGION"))
	out, err := client.DescribeVpcEndpoints(&ec2.DescribeVpcEndpointsInput{VpcEndpointIds: []*string{awsSDK.String(endpointId)}})
	require.NoError(t, err, "Failed to describe vpc endpoint %s", endpointId)
	require.Equal(t, 1, len(out.VpcEndpoints), "vpc endpoint %s not found", endpointId)
	return awsSDK.StringValue(out.VpcEndpoints[0].PolicyDocument)
}

func assertServicePolicy(t *testing.T, policy string, key string, action string) {
	assert.Contains(t, policy, action, "Wrong actions in %s endpoint policy", key)
	assert.Contains(t, policy, "aws:PrincipalAccount", "%s endpoint policy must only trust the current account", key)
	assert.NotRegexp(t, `"Action"\s*:\s*(\[\s*)?"\*"`, policy, "%s endpoint policy must not allow bare * action", key)
}
