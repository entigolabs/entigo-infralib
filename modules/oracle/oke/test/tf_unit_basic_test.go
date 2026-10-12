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
	"github.com/oracle/oci-go-sdk/v65/containerengine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerraformOke(t *testing.T) {
	t.Run("Biz", testTerraformOkeBiz)
	t.Run("Pri", testTerraformOkePri)
}

// biz.yaml only makes the API endpoint public; everything else is the module's default.
func testTerraformOkeBiz(t *testing.T) {
	t.Parallel()
	testTerraformOke(t, "biz", "10.96.0.0/16", true)
}

// pri.yaml also moves the service range off the default. pri has no kms, so its node disks get no key.
func testTerraformOkePri(t *testing.T) {
	t.Parallel()
	testTerraformOke(t, "pri", "10.97.0.0/16", false)
}

func testTerraformOke(t *testing.T, prefix string, servicesCidr string, withKms bool) {
	outputs := oracle.GetTFOutputs(t, prefix)
	// vpc and kms are in the net step, wherever this module is deployed.
	netOutputs := oracle.GetTFOutputsStep(t, prefix, "net")
	modulePrefix := fmt.Sprintf("%s-%s-oke", prefix, strings.ToLower(os.Getenv("STEP_NAME")))

	clusterId := tf.GetStringValue(t, outputs, "oke__cluster_id")
	require.NotEmpty(t, clusterId, "cluster_id was not returned")
	assert.Equal(t, modulePrefix, tf.GetStringValue(t, outputs, "oke__cluster_name"), "Wrong cluster_name returned")
	assert.True(t, strings.HasPrefix(tf.GetStringValue(t, outputs, "oke__kubernetes_version"), "v1.36."), "Wrong kubernetes_version returned")

	// With is_public_ip_enabled, kubeconfig points at the public endpoint.
	publicEndpoint := tf.GetStringValue(t, outputs, "oke__public_endpoint")
	require.NotEmpty(t, publicEndpoint, "public_endpoint was not returned")
	assert.Equal(t, "https://"+publicEndpoint, tf.GetStringValue(t, outputs, "oke__kubernetes_endpoint"), "kubernetes_endpoint is not the public endpoint")

	// The node count defaults: main and mon have none, so their pools are not created.
	assert.Empty(t, tf.GetStringValue(t, outputs, "oke__main_node_pool_id"), "main_node_pool_id should be empty when main has no nodes")
	assert.Empty(t, tf.GetStringValue(t, outputs, "oke__mon_node_pool_id"), "mon_node_pool_id should be empty when mon has no nodes")
	toolsNodePoolId := tf.GetStringValue(t, outputs, "oke__tools_node_pool_id")
	require.NotEmpty(t, toolsNodePoolId, "tools_node_pool_id was not returned")

	client := containerEngineClient(t)

	clusterResponse, err := client.GetCluster(context.Background(), containerengine.GetClusterRequest{ClusterId: &clusterId})
	require.NoError(t, err, "Failed to get cluster %s", clusterId)
	cluster := clusterResponse.Cluster
	assert.Equal(t, tf.GetStringValue(t, netOutputs, "vpc__vpc_id"), *cluster.VcnId, "Cluster is not in the vpc module's VCN")
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
		assert.Equal(t, tf.GetStringValue(t, netOutputs, "kms__data_key_id"), *pool.NodeConfigDetails.KmsKeyId, "Tools nodes are not encrypted with the data key")
	} else {
		assert.Nil(t, pool.NodeConfigDetails.KmsKeyId, "Tools nodes have a KMS key without kms deployed")
	}
	privateSubnets := tf.GetStringListValue(t, netOutputs, "vpc__private_subnets")
	for _, placement := range pool.NodeConfigDetails.PlacementConfigs {
		assert.Equal(t, privateSubnets[0], *placement.SubnetId, "Tools nodes are not in the first private subnet")
	}
}

func containerEngineClient(t *testing.T) containerengine.ContainerEngineClient {
	provider := ocicommon.DefaultConfigProvider()
	if configFile := os.Getenv("OCI_CONFIG_FILE"); configFile != "" {
		provider = ocicommon.CustomProfileConfigProvider(configFile, "DEFAULT")
	}
	client, err := containerengine.NewContainerEngineClientWithConfigurationProvider(provider)
	require.NoError(t, err, "Failed to create the Container Engine client")
	if region := os.Getenv("OCI_REGION"); region != "" {
		client.SetRegion(region)
	}
	return client
}
