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
	assert.Equal(t, []string{"10.146.160.0/24"}, tf.GetStringListValue(t, outputs, "vpc__agc_subnet_cidrs"))
	assert.Equal(t, []string{"10.146.161.0/26"}, tf.GetStringListValue(t, outputs, "vpc__apiserver_subnet_cidrs"))
	assert.Equal(t, []string{"10.146.161.64/26"}, tf.GetStringListValue(t, outputs, "vpc__container_apps_subnet_cidrs"))
	assert.Equal(t, map[string]interface{}{"mssql": "10.146.192.0/24"}, tf.GetValue(t, outputs, "vpc__custom_subnet_cidrs"))

	assert.Equal(t, 2, len(tf.GetStringListValue(t, outputs, "vpc__private_subnets")), "Wrong number of private_subnets returned")
	assert.NotEmpty(t, tf.GetValue(t, outputs, "vpc__custom_subnets").(map[string]interface{})["mssql"], "custom subnet mssql was not returned")
	assert.Equal(t, 1, len(tf.GetStringListValue(t, outputs, "vpc__container_apps_environment_ids")), "Wrong number of container_apps_environment_ids returned")
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "vpc__nat_gateway_id"), "nat_gateway_id was not returned")
	assert.NotEmpty(t, tf.GetStringListValue(t, outputs, "vpc__nat_public_ips"), "nat_public_ips was not returned")
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "vpc__vpc_flow_log_id"), "vpc_flow_log_id was not returned")
}

func testTerraformVpcPri(t *testing.T) {
	t.Parallel()
	outputs := azure.GetTFOutputs(t, "pri")

	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "vpc__vpc_id"), "vpc_id was not returned")
	assert.Equal(t, "10.24.0.0/16", tf.GetStringValue(t, outputs, "vpc__vpc_cidr"))

	assert.Equal(t, []string{"10.24.64.0/18"}, tf.GetStringListValue(t, outputs, "vpc__private_subnet_cidrs"))
	assert.Equal(t, []string{"10.24.0.0/19"}, tf.GetStringListValue(t, outputs, "vpc__public_subnet_cidrs"))
	assert.Equal(t, []string{"10.24.32.0/19"}, tf.GetStringListValue(t, outputs, "vpc__intra_subnet_cidrs"))
	assert.Equal(t, []string{"10.24.128.0/19"}, tf.GetStringListValue(t, outputs, "vpc__database_subnet_cidrs"))
	assert.Equal(t, []string{"10.24.160.0/24"}, tf.GetStringListValue(t, outputs, "vpc__agc_subnet_cidrs"))
	assert.Equal(t, []string{"10.24.161.0/26"}, tf.GetStringListValue(t, outputs, "vpc__apiserver_subnet_cidrs"))
	assert.Equal(t, []string{"10.24.161.64/26"}, tf.GetStringListValue(t, outputs, "vpc__container_apps_subnet_cidrs"))
	assert.Empty(t, tf.GetValue(t, outputs, "vpc__custom_subnets"))

	assert.Equal(t, 1, len(tf.GetStringListValue(t, outputs, "vpc__container_apps_environment_ids")), "Wrong number of container_apps_environment_ids returned")
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "vpc__nat_gateway_id"), "nat_gateway_id was not returned")
}

func testTerraformVpcSpoke(t *testing.T) {
	t.Parallel()
	outputs := azure.GetTFOutputs(t, "spoke")

	assert.Equal(t, "10.30.0.0/20", tf.GetStringValue(t, outputs, "vpc__vpc_cidr"))
	assert.Equal(t, []string{"10.30.4.0/22"}, tf.GetStringListValue(t, outputs, "vpc__private_subnet_cidrs"))
	assert.Equal(t, []string{"spoke-nodes"}, tf.GetStringListValue(t, outputs, "vpc__private_subnet_names"))
	assert.Equal(t, []string{"10.30.0.0/23"}, tf.GetStringListValue(t, outputs, "vpc__public_subnet_cidrs"))
	assert.Equal(t, []string{"10.30.2.0/23"}, tf.GetStringListValue(t, outputs, "vpc__intra_subnet_cidrs"))
	assert.Empty(t, tf.GetStringListValue(t, outputs, "vpc__database_subnet_cidrs"))
	assert.Equal(t, []string{"10.30.10.0/24"}, tf.GetStringListValue(t, outputs, "vpc__agc_subnet_cidrs"))
	assert.Equal(t, []string{"10.30.11.0/26"}, tf.GetStringListValue(t, outputs, "vpc__apiserver_subnet_cidrs"))
	assert.Empty(t, tf.GetStringListValue(t, outputs, "vpc__container_apps_subnet_cidrs"))
	assert.Empty(t, tf.GetStringListValue(t, outputs, "vpc__container_apps_environment_ids"))
	assert.Empty(t, tf.GetStringValue(t, outputs, "vpc__nat_gateway_id"), "nat_gateway_id must be empty without NAT")
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "vpc__egress_route_table_id"), "egress_route_table_id was not returned")
	assert.Empty(t, tf.GetStringListValue(t, outputs, "vpc__nat_public_ips"))
}
