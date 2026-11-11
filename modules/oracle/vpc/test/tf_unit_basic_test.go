package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/oracle"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

// The egress each subnet tier's route table gives it.
const (
	internet = "internet gateway"
	nat      = "NAT gateway"
	services = "service gateway"
)

type tier struct {
	name   string
	cidrs  []string
	public bool
	routes []string
}

func TestVpc(t *testing.T) {
	env.RunEach(t, map[string]env.TestFunc{
		"oracle_biz": testVpcBiz,
		"oracle_pri": testVpcPri,
	})
}

// oracle_biz.yaml lists every tier but intra and pod, which take the module's defaults.
func testVpcBiz(t *testing.T, e *env.Environment) {
	testVpc(t, e, "10.201.0.0/16", []tier{
		{name: "public", cidrs: []string{"10.201.0.0/20"}, public: true, routes: []string{internet}},
		{name: "private", cidrs: []string{"10.201.16.0/20", "10.201.32.0/20"}, routes: []string{nat, services}},
		{name: "pod", cidrs: []string{"10.201.64.0/18"}, routes: []string{nat, services}},
		{name: "intra", cidrs: []string{"10.201.160.0/19"}, routes: []string{services}},
		{name: "database", cidrs: []string{"10.201.48.0/22"}, routes: []string{services}},
	})
}

// oracle_pri.yaml gives only the CIDR, so every tier is the module's default sizing of a /18.
func testVpcPri(t *testing.T, e *env.Environment) {
	testVpc(t, e, "10.202.0.0/18", []tier{
		{name: "public", cidrs: []string{"10.202.32.0/21"}, public: true, routes: []string{internet}},
		{name: "private", cidrs: []string{"10.202.0.0/21"}, routes: []string{nat, services}},
		{name: "pod", cidrs: []string{"10.202.16.0/20"}, routes: []string{nat, services}},
		{name: "intra", cidrs: []string{"10.202.40.0/21"}, routes: []string{services}},
		{name: "database", cidrs: []string{"10.202.48.0/21"}, routes: []string{services}},
	})
}

func testVpc(t *testing.T, e *env.Environment, vpcCidr string, tiers []tier) {
	outputs := tf.Get(t, e)
	modulePrefix := fmt.Sprintf("%s-%s-vpc", e.Prefix, env.ModulePlacement(t, e).Step.Name)

	vpcId := outputs.String(t, "vpc__vpc_id")
	require.NotEmpty(t, vpcId, "vpc_id was not returned")
	assert.Equal(t, modulePrefix, outputs.String(t, "vpc__vpc_name"), "Wrong vpc_name returned")
	assert.Equal(t, vpcCidr, outputs.String(t, "vpc__vpc_cidr"), "Wrong vpc_cidr returned")

	// All three gateways are on by default.
	gateways := map[string]string{
		internet: outputs.String(t, "vpc__internet_gateway_id"),
		nat:      outputs.String(t, "vpc__nat_gateway_id"),
		services: outputs.String(t, "vpc__service_gateway_id"),
	}
	for name, id := range gateways {
		require.NotEmpty(t, id, "No %s returned", name)
	}

	client, err := core.NewVirtualNetworkClientWithConfigurationProvider(oracle.ConfigProvider())
	require.NoError(t, err, "Failed to create the Virtual Network client")
	client.SetRegion(e.MustRegion(t))

	vcnResponse, err := client.GetVcn(context.Background(), core.GetVcnRequest{VcnId: &vpcId})
	require.NoError(t, err, "Failed to get VCN %s", vpcId)
	assert.Equal(t, []string{vpcCidr}, vcnResponse.Vcn.CidrBlocks, "Wrong VCN CIDR blocks")

	for _, tier := range tiers {
		subnetIds := outputs.StringList(t, "vpc__"+tier.name+"_subnets")
		require.Len(t, subnetIds, len(tier.cidrs), "Wrong number of %s subnets", tier.name)
		assert.Equal(t, tier.cidrs, outputs.StringList(t, "vpc__"+tier.name+"_subnet_cidrs"), "Wrong %s_subnet_cidrs returned", tier.name)

		for i, subnetId := range subnetIds {
			response, err := client.GetSubnet(context.Background(), core.GetSubnetRequest{SubnetId: &subnetId})
			require.NoError(t, err, "Failed to get %s subnet %d", tier.name, i)
			subnet := response.Subnet
			assert.Equal(t, fmt.Sprintf("%s-%s-%d", modulePrefix, tier.name, i), *subnet.DisplayName, "Wrong name for %s subnet %d", tier.name, i)
			assert.Equal(t, vpcId, *subnet.VcnId, "%s subnet %d is not in the VCN", tier.name, i)
			assert.Equal(t, tier.cidrs[i], *subnet.CidrBlock, "Wrong CIDR for %s subnet %d", tier.name, i)
			assert.Equal(t, !tier.public, *subnet.ProhibitPublicIpOnVnic, "Wrong public IP setting for %s subnet %d", tier.name, i)
			assert.ElementsMatch(t, tier.routes, routeTargets(t, client, *subnet.RouteTableId, gateways), "Wrong routes for %s subnet %d", tier.name, i)
			// OKE requires a regional pod subnet.
			if tier.name == "pod" {
				assert.Nil(t, subnet.AvailabilityDomain, "Pod subnet %d is not regional", i)
			}
		}
	}

	// The singular aliases name the first subnet of their tier.
	assert.Equal(t, outputs.StringList(t, "vpc__public_subnets")[0], outputs.String(t, "vpc__public_subnet_id"), "public_subnet_id is not the first public subnet")
	assert.Equal(t, outputs.StringList(t, "vpc__private_subnets")[0], outputs.String(t, "vpc__private_subnet_id"), "private_subnet_id is not the first private subnet")
}

// The gateways a route table sends traffic to, by name. OCI does not keep the rule order.
func routeTargets(t *testing.T, client core.VirtualNetworkClient, routeTableId string, gateways map[string]string) []string {
	response, err := client.GetRouteTable(context.Background(), core.GetRouteTableRequest{RtId: &routeTableId})
	require.NoError(t, err, "Failed to get route table %s", routeTableId)
	targets := []string{}
	for _, rule := range response.RouteTable.RouteRules {
		target := *rule.NetworkEntityId
		for name, id := range gateways {
			if id == target {
				target = name
			}
		}
		targets = append(targets, target)
	}
	return targets
}
