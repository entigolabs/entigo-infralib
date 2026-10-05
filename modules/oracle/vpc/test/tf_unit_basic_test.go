package test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/oracle"
	"github.com/entigolabs/entigo-infralib-common/tf"
	ocicommon "github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestTerraformVpc(t *testing.T) {
	t.Run("Biz", testTerraformVpcBiz)
	t.Run("Pri", testTerraformVpcPri)
}

// biz.yaml lists every tier but intra and pod, which take the module's defaults.
func testTerraformVpcBiz(t *testing.T) {
	t.Parallel()
	testTerraformVpc(t, "biz", "10.201.0.0/16", []tier{
		{name: "public", cidrs: []string{"10.201.0.0/20"}, public: true, routes: []string{internet}},
		{name: "private", cidrs: []string{"10.201.16.0/20", "10.201.32.0/20"}, routes: []string{nat, services}},
		{name: "pod", cidrs: []string{"10.201.64.0/18"}, routes: []string{nat, services}},
		{name: "intra", cidrs: []string{"10.201.160.0/19"}, routes: []string{services}},
		{name: "database", cidrs: []string{"10.201.48.0/22"}, routes: []string{services}},
	})
}

// pri.yaml gives only the CIDR, so every tier is the module's default sizing of a /18.
func testTerraformVpcPri(t *testing.T) {
	t.Parallel()
	testTerraformVpc(t, "pri", "10.202.0.0/18", []tier{
		{name: "public", cidrs: []string{"10.202.32.0/21"}, public: true, routes: []string{internet}},
		{name: "private", cidrs: []string{"10.202.0.0/21"}, routes: []string{nat, services}},
		{name: "pod", cidrs: []string{"10.202.16.0/20"}, routes: []string{nat, services}},
		{name: "intra", cidrs: []string{"10.202.40.0/21"}, routes: []string{services}},
		{name: "database", cidrs: []string{"10.202.48.0/21"}, routes: []string{services}},
	})
}

func testTerraformVpc(t *testing.T, prefix string, vpcCidr string, tiers []tier) {
	outputs := oracle.GetTFOutputs(t, prefix)
	modulePrefix := fmt.Sprintf("%s-%s-vpc", prefix, strings.ToLower(os.Getenv("STEP_NAME")))

	vpcId := tf.GetStringValue(t, outputs, "vpc__vpc_id")
	require.NotEmpty(t, vpcId, "vpc_id was not returned")
	assert.Equal(t, modulePrefix, tf.GetStringValue(t, outputs, "vpc__vpc_name"), "Wrong vpc_name returned")
	assert.Equal(t, vpcCidr, tf.GetStringValue(t, outputs, "vpc__vpc_cidr"), "Wrong vpc_cidr returned")

	// All three gateways are on by default.
	gateways := map[string]string{
		internet: tf.GetStringValue(t, outputs, "vpc__internet_gateway_id"),
		nat:      tf.GetStringValue(t, outputs, "vpc__nat_gateway_id"),
		services: tf.GetStringValue(t, outputs, "vpc__service_gateway_id"),
	}
	for name, id := range gateways {
		require.NotEmpty(t, id, "No %s returned", name)
	}

	client, err := core.NewVirtualNetworkClientWithConfigurationProvider(configProvider())
	require.NoError(t, err, "Failed to create the Virtual Network client")
	if region := os.Getenv("OCI_REGION"); region != "" {
		client.SetRegion(region)
	}

	vcnResponse, err := client.GetVcn(context.Background(), core.GetVcnRequest{VcnId: &vpcId})
	require.NoError(t, err, "Failed to get VCN %s", vpcId)
	assert.Equal(t, []string{vpcCidr}, vcnResponse.Vcn.CidrBlocks, "Wrong VCN CIDR blocks")

	for _, tier := range tiers {
		subnetIds := tf.GetStringListValue(t, outputs, "vpc__"+tier.name+"_subnets")
		require.Len(t, subnetIds, len(tier.cidrs), "Wrong number of %s subnets", tier.name)
		assert.Equal(t, tier.cidrs, tf.GetStringListValue(t, outputs, "vpc__"+tier.name+"_subnet_cidrs"), "Wrong %s_subnet_cidrs returned", tier.name)

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
	assert.Equal(t, tf.GetStringListValue(t, outputs, "vpc__public_subnets")[0], tf.GetStringValue(t, outputs, "vpc__public_subnet_id"), "public_subnet_id is not the first public subnet")
	assert.Equal(t, tf.GetStringListValue(t, outputs, "vpc__private_subnets")[0], tf.GetStringValue(t, outputs, "vpc__private_subnet_id"), "private_subnet_id is not the first private subnet")
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

func configProvider() ocicommon.ConfigurationProvider {
	if configFile := os.Getenv("OCI_CONFIG_FILE"); configFile != "" {
		return ocicommon.CustomProfileConfigProvider(configFile, "DEFAULT")
	}
	return ocicommon.DefaultConfigProvider()
}
