package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/entigolabs/entigo-infralib-test/azure"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestVpc(t *testing.T) {
	env.RunEach(t, map[string]env.TestFunc{
		"azure_biz":   testVpcBiz,
		"azure_pri":   testVpcPri,
		"azure_spoke": testVpcSpoke,
		"azure_small": testVpcSmall,
	})
}

// <prefix>-<step>-vpc, the name the module gives its subnets.
func vpcPrefix(t *testing.T, e *env.Environment) string {
	return fmt.Sprintf("%s-%s-vpc", e.Prefix, strings.ToLower(env.ModulePlacement(t, e).Step.Name))
}

func testVpcBiz(t *testing.T, e *env.Environment) {
	outputs := tf.Get(t, e)

	assert.NotEmpty(t, outputs.String(t, "vpc__vpc_id"), "vpc_id was not returned")
	assert.Equal(t, outputs.String(t, "vpc__vpc_name"), outputs.String(t, "vpc__name"), "vpc_name and name differ")
	assert.Equal(t, "10.146.0.0/16", outputs.String(t, "vpc__vpc_cidr"))

	assert.Equal(t, []string{"10.146.32.0/21", "10.146.40.0/21"}, outputs.StringList(t, "vpc__private_subnet_cidrs"))
	assert.Equal(t, []string{"10.146.4.0/24", "10.146.5.0/24"}, outputs.StringList(t, "vpc__public_subnet_cidrs"))
	assert.Equal(t, []string{"10.146.16.0/22", "10.146.20.0/22"}, outputs.StringList(t, "vpc__database_subnet_cidrs"))
	assert.Empty(t, outputs.StringList(t, "vpc__intra_subnet_cidrs"))
	assert.Equal(t, []string{"10.146.160.0/24"}, outputs.StringList(t, "vpc__public_agc_subnet_cidrs"))
	assert.Equal(t, 1, len(outputs.StringList(t, "vpc__public_agc_subnets")), "Wrong number of agc_subnets returned")
	assert.Equal(t, []string{"10.146.162.0/26"}, outputs.StringList(t, "vpc__aks_subnet_cidrs"))
	assert.Equal(t, []string{"10.146.162.64/26"}, outputs.StringList(t, "vpc__container_apps_subnet_cidrs"))
	assert.Equal(t, []string{"10.146.161.0/24"}, outputs.StringList(t, "vpc__private_agc_subnet_cidrs"))
	assert.Equal(t, map[string]any{"custommssql": "10.146.192.0/24"}, outputs.Map(t, "vpc__custom_subnet_cidrs"))

	assert.Equal(t, 2, len(outputs.StringList(t, "vpc__private_subnets")), "Wrong number of private_subnets returned")
	assert.NotEmpty(t, outputs.Map(t, "vpc__custom_subnets")["custommssql"], "custom subnet custommssql was not returned")
	assert.Equal(t, 1, len(outputs.StringList(t, "vpc__container_apps_environment_ids")), "Wrong number of container_apps_environment_ids returned")
	assert.NotEmpty(t, outputs.String(t, "vpc__nat_gateway_id"), "nat_gateway_id was not returned")
	assert.NotEmpty(t, outputs.StringList(t, "vpc__nat_public_ips"), "nat_public_ips was not returned")
	assert.NotEmpty(t, outputs.String(t, "vpc__vpc_flow_log_id"), "vpc_flow_log_id was not returned")
}

func testVpcPri(t *testing.T, e *env.Environment) {
	outputs := tf.Get(t, e)

	assert.NotEmpty(t, outputs.String(t, "vpc__vpc_id"), "vpc_id was not returned")
	assert.Equal(t, "10.24.0.0/16", outputs.String(t, "vpc__vpc_cidr"))

	assert.Equal(t, []string{"10.24.64.0/18"}, outputs.StringList(t, "vpc__private_subnet_cidrs"))
	assert.Equal(t, []string{"10.24.0.0/25"}, outputs.StringList(t, "vpc__public_subnet_cidrs"))
	assert.Equal(t, []string{"10.24.0.128/25"}, outputs.StringList(t, "vpc__intra_subnet_cidrs"))
	assert.Equal(t, []string{"10.24.128.0/19"}, outputs.StringList(t, "vpc__database_subnet_cidrs"))
	assert.Equal(t, []string{"10.24.160.0/24"}, outputs.StringList(t, "vpc__public_agc_subnet_cidrs"))
	assert.Equal(t, []string{"10.24.162.0/26"}, outputs.StringList(t, "vpc__aks_subnet_cidrs"))
	assert.Equal(t, []string{"10.24.162.64/26"}, outputs.StringList(t, "vpc__container_apps_subnet_cidrs"))
	prefix := vpcPrefix(t, e)
	assert.True(t, strings.HasSuffix(outputs.StringList(t, "vpc__aks_subnets")[0], "/subnets/"+prefix+"-aks-0"), "aks subnet must be named <prefix>-aks-0")
	assert.True(t, strings.HasSuffix(outputs.StringList(t, "vpc__public_agc_subnets")[0], "/subnets/"+prefix+"-public-agc-0"), "AGC subnet must be named <prefix>-public-agc-0")
	assert.Equal(t, []string{"10.24.161.0/24"}, outputs.StringList(t, "vpc__private_agc_subnet_cidrs"))
	assert.True(t, strings.HasSuffix(outputs.StringList(t, "vpc__private_agc_subnets")[0], "/subnets/"+prefix+"-private-agc-0"), "internal AGC subnet must be named <prefix>-private-agc-0")

	assert.Equal(t, 1, len(outputs.StringList(t, "vpc__container_apps_environment_ids")), "Wrong number of container_apps_environment_ids returned")
	assert.NotEmpty(t, outputs.String(t, "vpc__nat_gateway_id"), "nat_gateway_id was not returned")
	assert.Empty(t, outputs.StringList(t, "vpc__control_agc_subnets"), "no control_agc_subnets in default mode")
	assert.Empty(t, outputs.StringList(t, "vpc__service_agc_subnets"), "no service_agc_subnets in default mode")
	assert.Empty(t, outputs.StringList(t, "vpc__control_subnets"), "no control_subnets in default mode")
}

// /21 spoke without the managed AGC subnet (edge in the hub), explicit intra and control/service AGC subnet names, Standard NAT with two IPs, custom
// subnets in the free AGC /24: SQL managed instance, VMs with NAT egress
func testVpcSpoke(t *testing.T, e *env.Environment) {
	outputs := tf.Get(t, e)

	assert.Equal(t, "10.33.0.0/21", outputs.String(t, "vpc__vpc_cidr"))
	assert.Equal(t, []string{"10.33.0.0/26"}, outputs.StringList(t, "vpc__public_subnet_cidrs"))
	assert.Equal(t, []string{"10.33.2.0/24"}, outputs.StringList(t, "vpc__control_agc_subnets_cidr_blocks"))
	assert.Equal(t, []string{"10.33.0.128/27"}, outputs.StringList(t, "vpc__aks_subnet_cidrs"))
	assert.Equal(t, []string{"10.33.3.0/24"}, outputs.StringList(t, "vpc__service_agc_subnets_cidr_blocks"))
	assert.Equal(t, []string{"10.33.0.64/26"}, outputs.StringList(t, "vpc__database_subnet_cidrs"))
	assert.Equal(t, []string{"10.33.0.192/26"}, outputs.StringList(t, "vpc__intra_subnet_cidrs"))
	assert.Equal(t, []string{"10.33.4.0/22"}, outputs.StringList(t, "vpc__private_subnet_cidrs"))
	assert.Equal(t, []string{"10.33.4.0/22"}, outputs.StringList(t, "vpc__compute_subnets_cidr_blocks"))
	assert.Equal(t, vpcPrefix(t, e)+"-compute-0", outputs.StringList(t, "vpc__private_subnet_names")[0])
	assert.Empty(t, outputs.StringList(t, "vpc__public_agc_subnets"))
	assert.Empty(t, outputs.StringList(t, "vpc__private_agc_subnets"), "no private-agc subnet in spoke mode")
	assert.Empty(t, outputs.StringList(t, "vpc__container_apps_subnet_cidrs"))
	assert.Empty(t, outputs.StringList(t, "vpc__container_apps_environment_ids"))
	assert.Equal(t, []string{"spoke-agc-control"}, outputs.StringList(t, "vpc__control_agc_subnet_names"))
	assert.Equal(t, []string{"spoke-agc-service"}, outputs.StringList(t, "vpc__service_agc_subnet_names"))
	assert.Equal(t, []string{"10.33.1.64/28"}, outputs.StringList(t, "vpc__control_subnets_cidr_blocks"))
	assert.Equal(t, []string{"10.33.1.80/28"}, outputs.StringList(t, "vpc__service_subnets_cidr_blocks"))
	assert.NotEqual(t, outputs.StringList(t, "vpc__private_subnets"), outputs.StringList(t, "vpc__control_subnets"), "control subnet must not be the node subnet in spoke mode")
	assert.NotEqual(t, outputs.StringList(t, "vpc__private_subnets"), outputs.StringList(t, "vpc__control_agc_subnets"), "control-agc subnet must not be the node subnet in spoke mode")
	assert.Equal(t, 2, len(outputs.StringList(t, "vpc__nat_public_ips")), "Wrong number of NAT public IPs")
	assert.NotEmpty(t, outputs.String(t, "vpc__nat_gateway_id"), "nat_gateway_id was not returned")

	custom := outputs.Map(t, "vpc__custom_subnets")
	assert.NotEmpty(t, custom["custommssql"], "custom subnet custommssql was not returned")
	assert.NotEmpty(t, custom["customvms"], "custom subnet customvms was not returned")
	assert.Equal(t, map[string]any{"custommssql": "10.33.1.0/27", "customvms": "10.33.1.128/26"}, outputs.Map(t, "vpc__custom_subnet_cidrs"))
	assert.Empty(t, outputs.String(t, "vpc__vpc_flow_log_id"))
}

// /21 default mode: the automatic small layout with firewall egress instead of NAT
func testVpcSmall(t *testing.T, e *env.Environment) {
	outputs := tf.Get(t, e)

	assert.Equal(t, []string{"10.35.0.0/26"}, outputs.StringList(t, "vpc__public_subnet_cidrs"))
	assert.Equal(t, []string{"10.35.0.64/26"}, outputs.StringList(t, "vpc__intra_subnet_cidrs"))
	assert.Equal(t, []string{"10.35.0.128/27"}, outputs.StringList(t, "vpc__aks_subnet_cidrs"))
	assert.Empty(t, outputs.StringList(t, "vpc__database_subnet_cidrs"), "no database subnet in a default /21")
	assert.Equal(t, []string{"10.35.1.0/24"}, outputs.StringList(t, "vpc__public_agc_subnet_cidrs"))
	assert.Equal(t, []string{"10.35.2.0/24"}, outputs.StringList(t, "vpc__private_agc_subnet_cidrs"))
	assert.Equal(t, []string{"10.35.4.0/22"}, outputs.StringList(t, "vpc__private_subnet_cidrs"))
	assert.Empty(t, outputs.StringList(t, "vpc__container_apps_subnet_cidrs"))
	assert.Empty(t, outputs.String(t, "vpc__nat_gateway_id"), "nat_gateway_id must be empty without NAT")
	assert.Empty(t, outputs.StringList(t, "vpc__nat_public_ips"))
	assert.NotEmpty(t, outputs.String(t, "vpc__egress_route_table_id"), "egress_route_table_id was not returned")
	assert.Equal(t, outputs.StringList(t, "vpc__private_subnets"), outputs.StringList(t, "vpc__compute_subnets"))
	assert.Empty(t, outputs.StringList(t, "vpc__control_agc_subnets"), "no control_agc_subnets in default mode")
	assert.Empty(t, outputs.Value(t, "vpc__custom_subnets"))
}
