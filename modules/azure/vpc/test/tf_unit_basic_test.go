package test

import (
	"testing"

	"github.com/entigolabs/entigo-infralib-common/azure"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestTerraformVpc(t *testing.T) {
	t.Run("Biz", testTerraformVpcBiz)
	t.Run("Pri", testTerraformVpcPri)
	t.Run("Spoke", testTerraformVpcSpoke)
}

func testTerraformVpcBiz(t *testing.T) {
	t.Parallel()
	outputs := azure.GetTFOutputs(t, "biz")

	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "vpc__vpc_id"), "vpc_id was not returned")
	assert.Equal(t, tf.GetStringValue(t, outputs, "vpc__vpc_name"), tf.GetStringValue(t, outputs, "vpc__name"), "vpc_name and name differ")
	assert.Equal(t, "10.146.0.0/16", tf.GetStringValue(t, outputs, "vpc__vpc_cidr"))

	assert.Equal(t, []string{"10.146.32.0/21", "10.146.40.0/21"}, tf.GetStringListValue(t, outputs, "vpc__private_subnet_cidrs"))
	assert.Equal(t, []string{"10.146.4.0/24", "10.146.5.0/24"}, tf.GetStringListValue(t, outputs, "vpc__public_subnet_cidrs"))
	assert.Equal(t, []string{"10.146.16.0/22", "10.146.20.0/22"}, tf.GetStringListValue(t, outputs, "vpc__database_subnet_cidrs"))
	assert.Empty(t, tf.GetStringListValue(t, outputs, "vpc__intra_subnet_cidrs"))
	assert.Equal(t, []string{"10.146.224.0/24"}, tf.GetStringListValue(t, outputs, "vpc__agc_subnet_cidrs"))
	assert.Equal(t, []string{"10.146.225.0/24"}, tf.GetStringListValue(t, outputs, "vpc__apiserver_subnet_cidrs"))
	assert.Equal(t, []string{"10.146.226.0/24"}, tf.GetStringListValue(t, outputs, "vpc__pipeline_subnet_cidrs"))
	assert.Equal(t, []string{"10.146.248.0/21"}, tf.GetStringListValue(t, outputs, "vpc__mssql_subnet_cidrs"))

	assert.Equal(t, 2, len(tf.GetStringListValue(t, outputs, "vpc__private_subnets")), "Wrong number of private_subnets returned")
	assert.Equal(t, 1, len(tf.GetStringListValue(t, outputs, "vpc__mssql_subnets")), "Wrong number of mssql_subnets returned")
	assert.Equal(t, 1, len(tf.GetStringListValue(t, outputs, "vpc__pipeline_environment_ids")), "Wrong number of pipeline_environment_ids returned")
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "vpc__nat_gateway_id"), "nat_gateway_id was not returned")
	assert.NotEmpty(t, tf.GetStringListValue(t, outputs, "vpc__nat_public_ips"), "nat_public_ips was not returned")
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "vpc__vpc_flow_log_id"), "vpc_flow_log_id was not returned")
}

func testTerraformVpcPri(t *testing.T) {
	t.Parallel()
	outputs := azure.GetTFOutputs(t, "pri")

	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "vpc__vpc_id"), "vpc_id was not returned")
	assert.Equal(t, "10.24.0.0/16", tf.GetStringValue(t, outputs, "vpc__vpc_cidr"))

	assert.Equal(t, []string{"10.24.0.0/17"}, tf.GetStringListValue(t, outputs, "vpc__private_subnet_cidrs"))
	assert.Equal(t, []string{"10.24.128.0/19"}, tf.GetStringListValue(t, outputs, "vpc__public_subnet_cidrs"))
	assert.Equal(t, []string{"10.24.160.0/19"}, tf.GetStringListValue(t, outputs, "vpc__intra_subnet_cidrs"))
	assert.Equal(t, []string{"10.24.192.0/19"}, tf.GetStringListValue(t, outputs, "vpc__database_subnet_cidrs"))
	assert.Equal(t, []string{"10.24.224.0/24"}, tf.GetStringListValue(t, outputs, "vpc__agc_subnet_cidrs"))
	assert.Equal(t, []string{"10.24.225.0/24"}, tf.GetStringListValue(t, outputs, "vpc__apiserver_subnet_cidrs"))
	assert.Equal(t, []string{"10.24.226.0/24"}, tf.GetStringListValue(t, outputs, "vpc__pipeline_subnet_cidrs"))
	assert.Empty(t, tf.GetStringListValue(t, outputs, "vpc__mssql_subnet_cidrs"))

	assert.Equal(t, 1, len(tf.GetStringListValue(t, outputs, "vpc__pipeline_environment_ids")), "Wrong number of pipeline_environment_ids returned")
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "vpc__nat_gateway_id"), "nat_gateway_id was not returned")
}

func testTerraformVpcSpoke(t *testing.T) {
	t.Parallel()
	outputs := azure.GetTFOutputs(t, "spoke")

	assert.Equal(t, "10.30.0.0/20", tf.GetStringValue(t, outputs, "vpc__vpc_cidr"))
	assert.Equal(t, []string{"10.30.0.0/21"}, tf.GetStringListValue(t, outputs, "vpc__private_subnet_cidrs"))
	assert.Equal(t, []string{"spoke-nodes"}, tf.GetStringListValue(t, outputs, "vpc__private_subnet_names"))
	assert.Equal(t, []string{"10.30.8.0/23"}, tf.GetStringListValue(t, outputs, "vpc__public_subnet_cidrs"))
	assert.Equal(t, []string{"10.30.10.0/23"}, tf.GetStringListValue(t, outputs, "vpc__intra_subnet_cidrs"))
	assert.Equal(t, []string{"10.30.12.0/23"}, tf.GetStringListValue(t, outputs, "vpc__database_subnet_cidrs"))
	assert.Equal(t, []string{"10.30.14.0/24"}, tf.GetStringListValue(t, outputs, "vpc__agc_subnet_cidrs"))
	assert.Equal(t, []string{"10.30.15.0/28"}, tf.GetStringListValue(t, outputs, "vpc__apiserver_subnet_cidrs"))
	assert.Empty(t, tf.GetStringListValue(t, outputs, "vpc__pipeline_subnet_cidrs"))
	assert.Empty(t, tf.GetStringListValue(t, outputs, "vpc__pipeline_environment_ids"))
	assert.False(t, tf.HasKeyWithPrefix(t, outputs, "vpc__nat_gateway_id"), "nat_gateway_id must be null without NAT")
	assert.Empty(t, tf.GetStringListValue(t, outputs, "vpc__nat_public_ips"))
}
