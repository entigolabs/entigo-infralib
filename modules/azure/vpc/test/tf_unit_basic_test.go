package test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/azure"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestTerraformVpc(t *testing.T) {
	t.Run("Biz", testTerraformVpcBiz)
	t.Run("Pri", testTerraformVpcPri)
	t.Run("Spoke", testTerraformVpcSpoke)
	t.Run("Small", testTerraformVpcSmall)
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
	assert.Equal(t, []string{"10.146.160.0/24", "10.146.193.0/24"}, tf.GetStringListValue(t, outputs, "vpc__public_agc_subnet_cidrs"))
	assert.Equal(t, 2, len(tf.GetStringListValue(t, outputs, "vpc__public_agc_subnets")), "Wrong number of public_agc_subnets returned")
	assert.Empty(t, tf.GetStringValue(t, outputs, "vpc__pod_subnet_id"), "biz is Overlay, no pod subnet")
	assert.Equal(t, []string{"10.146.161.0/26"}, tf.GetStringListValue(t, outputs, "vpc__aks_subnet_cidrs"))
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
	assert.Equal(t, []string{"10.24.160.0/24"}, tf.GetStringListValue(t, outputs, "vpc__public_agc_subnet_cidrs"))
	assert.Equal(t, []string{"10.24.161.0/26"}, tf.GetStringListValue(t, outputs, "vpc__aks_subnet_cidrs"))
	assert.Equal(t, []string{"10.24.161.64/26"}, tf.GetStringListValue(t, outputs, "vpc__container_apps_subnet_cidrs"))
	assert.Equal(t, []string{"10.24.192.0/19"}, tf.GetStringListValue(t, outputs, "vpc__pod_subnet_cidrs"), "enable_pod_subnets takes the first half of the spare quarter")
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "vpc__pod_subnet_id"), "pod_subnet_id was not returned")
	assert.Equal(t, tf.GetStringListValue(t, outputs, "vpc__pod_subnets")[0], tf.GetStringValue(t, outputs, "vpc__pod_subnet_id"))
	assert.Empty(t, tf.GetValue(t, outputs, "vpc__custom_subnets"))

	assert.Equal(t, 1, len(tf.GetStringListValue(t, outputs, "vpc__container_apps_environment_ids")), "Wrong number of container_apps_environment_ids returned")
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "vpc__nat_gateway_id"), "nat_gateway_id was not returned")
	assert.Equal(t, tf.GetStringListValue(t, outputs, "vpc__private_subnets"), tf.GetStringListValue(t, outputs, "vpc__control_subnets"), "control_subnets must be the private subnets in default mode")
	assert.Equal(t, tf.GetStringListValue(t, outputs, "vpc__private_subnets"), tf.GetStringListValue(t, outputs, "vpc__service_subnets"), "service_subnets must be the private subnets in default mode")
	assert.Empty(t, tf.GetStringListValue(t, outputs, "vpc__lb_subnets"), "no lb_subnets in default mode")
	assert.Empty(t, tf.GetStringValue(t, outputs, "vpc__control_subnet_name"), "no control_subnet_name in default mode")
}

// /21 spoke without AGC (edge in the hub), explicit intra and load balancer subnet names, Standard NAT with two IPs, custom
// subnets in the free AGC /24: SQL managed instance, DNS resolver, Bastion, VMs with NAT egress
func testTerraformVpcSpoke(t *testing.T) {
	t.Parallel()
	outputs := azure.GetTFOutputs(t, "spoke")

	assert.Equal(t, "10.33.0.0/21", tf.GetStringValue(t, outputs, "vpc__vpc_cidr"))
	assert.Equal(t, []string{"10.33.0.0/26"}, tf.GetStringListValue(t, outputs, "vpc__public_subnet_cidrs"))
	assert.Equal(t, []string{"10.33.0.64/26"}, tf.GetStringListValue(t, outputs, "vpc__control_subnets_cidr_blocks"))
	assert.Equal(t, []string{"10.33.0.128/28"}, tf.GetStringListValue(t, outputs, "vpc__aks_subnet_cidrs"))
	assert.Equal(t, []string{"10.33.0.192/26"}, tf.GetStringListValue(t, outputs, "vpc__service_subnets_cidr_blocks"))
	assert.Equal(t, []string{"10.33.2.0/24"}, tf.GetStringListValue(t, outputs, "vpc__database_subnet_cidrs"))
	assert.Equal(t, []string{"10.33.3.0/26"}, tf.GetStringListValue(t, outputs, "vpc__intra_subnet_cidrs"))
	assert.Equal(t, []string{"10.33.4.0/22"}, tf.GetStringListValue(t, outputs, "vpc__private_subnet_cidrs"))
	assert.Equal(t, []string{"10.33.4.0/22"}, tf.GetStringListValue(t, outputs, "vpc__compute_subnets_cidr_blocks"))
	assert.Equal(t, fmt.Sprintf("spoke-%s-vpc-compute-0", strings.ToLower(os.Getenv("STEP_NAME"))), tf.GetStringListValue(t, outputs, "vpc__private_subnet_names")[0])
	assert.Empty(t, tf.GetStringListValue(t, outputs, "vpc__public_agc_subnets"))
	assert.Empty(t, tf.GetStringListValue(t, outputs, "vpc__container_apps_subnet_cidrs"))
	assert.Empty(t, tf.GetStringListValue(t, outputs, "vpc__container_apps_environment_ids"))
	assert.Equal(t, []string{"spoke-ilb-platform"}, tf.GetStringListValue(t, outputs, "vpc__control_subnet_names"))
	assert.Equal(t, "spoke-ilb-platform", tf.GetStringValue(t, outputs, "vpc__control_subnet_name"))
	assert.Equal(t, []string{"spoke-ilb-apps"}, tf.GetStringListValue(t, outputs, "vpc__service_subnet_names"))
	assert.Equal(t, 2, len(tf.GetStringListValue(t, outputs, "vpc__lb_subnets")), "lb_subnets must be control + service")
	assert.NotEqual(t, tf.GetStringListValue(t, outputs, "vpc__private_subnets"), tf.GetStringListValue(t, outputs, "vpc__control_subnets"), "control subnet must not be the node subnet in spoke mode")
	assert.Equal(t, 2, len(tf.GetStringListValue(t, outputs, "vpc__nat_public_ips")), "Wrong number of NAT public IPs")
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "vpc__nat_gateway_id"), "nat_gateway_id was not returned")

	custom := tf.GetValue(t, outputs, "vpc__custom_subnets").(map[string]interface{})
	assert.NotEmpty(t, custom["mssql"], "custom subnet mssql was not returned")
	assert.NotEmpty(t, custom["resolver"], "custom subnet resolver was not returned")
	assert.NotEmpty(t, custom["bastion"], "custom subnet bastion was not returned")
	assert.NotEmpty(t, custom["vms"], "custom subnet vms was not returned")
	assert.Equal(t, map[string]interface{}{"mssql": "10.33.1.0/27", "resolver": "10.33.1.32/28", "bastion": "10.33.1.64/26", "vms": "10.33.1.128/26"}, tf.GetValue(t, outputs, "vpc__custom_subnet_cidrs"))
	assert.Empty(t, tf.GetStringValue(t, outputs, "vpc__vpc_flow_log_id"))
}

// /21 default mode: the small layout with firewall egress instead of NAT, a second aks subnet and an explicit
// database subnet from the spare
func testTerraformVpcSmall(t *testing.T) {
	t.Parallel()
	outputs := azure.GetTFOutputs(t, "small")

	assert.Equal(t, []string{"10.35.0.0/26"}, tf.GetStringListValue(t, outputs, "vpc__public_subnet_cidrs"))
	assert.Equal(t, []string{"10.35.0.64/26"}, tf.GetStringListValue(t, outputs, "vpc__intra_subnet_cidrs"))
	assert.Equal(t, []string{"10.35.0.128/28", "10.35.0.144/28"}, tf.GetStringListValue(t, outputs, "vpc__aks_subnet_cidrs"))
	assert.Equal(t, 2, len(tf.GetStringListValue(t, outputs, "vpc__aks_subnets")), "Wrong number of aks_subnets returned")
	assert.Equal(t, []string{"10.35.0.192/26"}, tf.GetStringListValue(t, outputs, "vpc__database_subnet_cidrs"))
	assert.Equal(t, []string{"10.35.1.0/24"}, tf.GetStringListValue(t, outputs, "vpc__public_agc_subnet_cidrs"))
	assert.Equal(t, []string{"10.35.4.0/22"}, tf.GetStringListValue(t, outputs, "vpc__private_subnet_cidrs"))
	assert.Empty(t, tf.GetStringListValue(t, outputs, "vpc__container_apps_subnet_cidrs"))
	assert.Empty(t, tf.GetStringValue(t, outputs, "vpc__nat_gateway_id"), "nat_gateway_id must be empty without NAT")
	assert.Empty(t, tf.GetStringListValue(t, outputs, "vpc__nat_public_ips"))
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "vpc__egress_route_table_id"), "egress_route_table_id was not returned")
	assert.Equal(t, tf.GetStringListValue(t, outputs, "vpc__private_subnets"), tf.GetStringListValue(t, outputs, "vpc__compute_subnets"))
	assert.Equal(t, tf.GetStringListValue(t, outputs, "vpc__private_subnets"), tf.GetStringListValue(t, outputs, "vpc__control_subnets"), "control_subnets must be the private subnets in default mode")
	assert.Empty(t, tf.GetStringListValue(t, outputs, "vpc__lb_subnets"))
	assert.Empty(t, tf.GetValue(t, outputs, "vpc__custom_subnets"))
}
