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

func TestTerraformOkeNodePool(t *testing.T) {
	t.Run("Biz", testTerraformOkeNodePoolBiz)
	t.Run("Pri", testTerraformOkeNodePoolPri)
}

func testTerraformOkeNodePoolBiz(t *testing.T) {
	t.Parallel()
	testTerraformOkeNodePool(t, "biz")
}

func testTerraformOkeNodePoolPri(t *testing.T) {
	t.Parallel()
	testTerraformOkeNodePool(t, "pri")
}

// Both test configurations set node_count to 1.
func testTerraformOkeNodePool(t *testing.T, prefix string) {
	outputs := oracle.GetTFOutputs(t, prefix)

	nodePoolId := tf.GetStringValue(t, outputs, "oke-node-pool__node_pool_id")
	require.NotEmpty(t, nodePoolId, "node_pool_id was not returned")

	nodePoolName := tf.GetStringValue(t, outputs, "oke-node-pool__node_pool_name")
	assert.Equal(t, fmt.Sprintf("%s-%s-oke-node-pool", prefix, strings.ToLower(os.Getenv("STEP_NAME"))), nodePoolName, "Wrong node_pool_name returned")

	// node_count is not an output, so the pool itself is read.
	nodePool := getNodePool(t, nodePoolId)
	assert.Equal(t, 1, *nodePool.NodeConfigDetails.Size, "Wrong node pool size")
	// The cluster comes from the infra step: a pull request deploys this module in a step of its own.
	clusterId := tf.GetStringValue(t, oracle.GetTFOutputsStep(t, prefix, "infra"), "oke__cluster_id")
	assert.Equal(t, clusterId, *nodePool.ClusterId, "Node pool is not in the infra step's cluster")
}

func getNodePool(t *testing.T, nodePoolId string) containerengine.NodePool {
	provider := ocicommon.DefaultConfigProvider()
	if configFile := os.Getenv("OCI_CONFIG_FILE"); configFile != "" {
		provider = ocicommon.CustomProfileConfigProvider(configFile, "DEFAULT")
	}
	client, err := containerengine.NewContainerEngineClientWithConfigurationProvider(provider)
	require.NoError(t, err, "Failed to create the Container Engine client")
	if region := os.Getenv("OCI_REGION"); region != "" {
		client.SetRegion(region)
	}
	response, err := client.GetNodePool(context.Background(), containerengine.GetNodePoolRequest{NodePoolId: &nodePoolId})
	require.NoError(t, err, "Failed to get node pool %s", nodePoolId)
	return response.NodePool
}
