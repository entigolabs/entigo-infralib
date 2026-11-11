package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/oracle/oci-go-sdk/v65/containerengine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/oracle"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestOke(t *testing.T) {
	env.RunEach(t, map[string]env.TestFunc{
		"oracle_biz": testOkeBiz,
		"oracle_pri": testOkePri,
	})
}

// oracle_biz.yaml only makes the API endpoint public; everything else is the module's default.
func testOkeBiz(t *testing.T, e *env.Environment) {
	testOke(t, e, "10.96.0.0/16", true)
}

// oracle_pri.yaml also moves the service range off the default. pri has no kms, so its node disks get no key.
func testOkePri(t *testing.T, e *env.Environment) {
	testOke(t, e, "10.97.0.0/16", false)
}

func testOke(t *testing.T, e *env.Environment, servicesCidr string, withKms bool) {
	outputs := tf.Get(t, e)
	// vpc and kms are in the net step, wherever this module is deployed.
	netOutputs := tf.GetStep(t, e, "net")
	modulePrefix := fmt.Sprintf("%s-%s-oke", e.Prefix, env.ModulePlacement(t, e).Step.Name)

	clusterId := outputs.String(t, "oke__cluster_id")
	require.NotEmpty(t, clusterId, "cluster_id was not returned")
	assert.Equal(t, modulePrefix, outputs.String(t, "oke__cluster_name"), "Wrong cluster_name returned")
	assert.True(t, strings.HasPrefix(outputs.String(t, "oke__kubernetes_version"), "v1.36."), "Wrong kubernetes_version returned")

	// With is_public_ip_enabled, kubeconfig points at the public endpoint.
	publicEndpoint := outputs.String(t, "oke__public_endpoint")
	require.NotEmpty(t, publicEndpoint, "public_endpoint was not returned")
	assert.Equal(t, "https://"+publicEndpoint, outputs.String(t, "oke__kubernetes_endpoint"), "kubernetes_endpoint is not the public endpoint")

	// The node count defaults: main and mon have none, so their pools are not created.
	assert.Empty(t, outputs.String(t, "oke__main_node_pool_id"), "main_node_pool_id should be empty when main has no nodes")
	assert.Empty(t, outputs.String(t, "oke__mon_node_pool_id"), "mon_node_pool_id should be empty when mon has no nodes")
	toolsNodePoolId := outputs.String(t, "oke__tools_node_pool_id")
	require.NotEmpty(t, toolsNodePoolId, "tools_node_pool_id was not returned")

	client := containerEngineClient(t, e)

	clusterResponse, err := client.GetCluster(context.Background(), containerengine.GetClusterRequest{ClusterId: &clusterId})
	require.NoError(t, err, "Failed to get cluster %s", clusterId)
	cluster := clusterResponse.Cluster
	assert.Equal(t, netOutputs.String(t, "vpc__vpc_id"), *cluster.VcnId, "Cluster is not in the vpc module's VCN")
	assert.True(t, *cluster.EndpointConfig.IsPublicIpEnabled, "Cluster API endpoint is not public")
	assert.Equal(t, servicesCidr, *cluster.Options.KubernetesNetworkConfig.ServicesCidr, "Wrong services CIDR")
	assert.Equal(t, "10.244.0.0/16", *cluster.Options.KubernetesNetworkConfig.PodsCidr, "Wrong pods CIDR")

	// The tools pool takes the module's tools defaults, in the private subnet.
	poolResponse, err := client.GetNodePool(context.Background(), containerengine.GetNodePoolRequest{NodePoolId: &toolsNodePoolId})
	require.NoError(t, err, "Failed to get node pool %s", toolsNodePoolId)
	pool := poolResponse.NodePool
	assert.Equal(t, modulePrefix+"-tools", *pool.Name, "Wrong tools node pool name")
	assert.Equal(t, clusterId, *pool.ClusterId, "Tools node pool is not in the cluster")
	assert.Equal(t, 2, *pool.NodeConfigDetails.Size, "Wrong tools node pool size")
	assert.Equal(t, "VM.Standard.E4.Flex", *pool.NodeShape, "Wrong tools node shape")
	assert.Equal(t, float32(2), *pool.NodeShapeConfig.Ocpus, "Wrong tools node OCPUs")
	assert.Equal(t, float32(8), *pool.NodeShapeConfig.MemoryInGBs, "Wrong tools node memory")
	if withKms {
		require.NotNil(t, pool.NodeConfigDetails.KmsKeyId, "Tools nodes have no KMS key")
		assert.Equal(t, netOutputs.String(t, "kms__data_key_id"), *pool.NodeConfigDetails.KmsKeyId, "Tools nodes are not encrypted with the data key")
	} else {
		assert.Nil(t, pool.NodeConfigDetails.KmsKeyId, "Tools nodes have a KMS key without kms deployed")
	}
	privateSubnets := netOutputs.StringList(t, "vpc__private_subnets")
	for _, placement := range pool.NodeConfigDetails.PlacementConfigs {
		assert.Equal(t, privateSubnets[0], *placement.SubnetId, "Tools nodes are not in the first private subnet")
	}
}

func containerEngineClient(t *testing.T, e *env.Environment) containerengine.ContainerEngineClient {
	client, err := containerengine.NewContainerEngineClientWithConfigurationProvider(oracle.ConfigProvider())
	require.NoError(t, err, "Failed to create the Container Engine client")
	client.SetRegion(e.MustRegion(t))
	return client
}
