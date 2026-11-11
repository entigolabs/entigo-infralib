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

// Both test configurations set node_count to 1.
func TestOkeNodePool(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)

		nodePoolId := outputs.String(t, "oke-node-pool__node_pool_id")
		require.NotEmpty(t, nodePoolId, "node_pool_id was not returned")

		// The module truncates the name to OCI's 32-character limit, dropping any trailing hyphen.
		expectedName := fmt.Sprintf("%s-%s-oke-node-pool", e.Prefix, env.ModulePlacement(t, e).Step.Name)
		if len(expectedName) > 32 {
			expectedName = strings.TrimRight(expectedName[:32], "-")
		}
		nodePoolName := outputs.String(t, "oke-node-pool__node_pool_name")
		assert.Equal(t, expectedName, nodePoolName, "Wrong node_pool_name returned")

		// node_count is not an output, so the pool itself is read.
		nodePool := getNodePool(t, e, nodePoolId)
		assert.Equal(t, 1, *nodePool.NodeConfigDetails.Size, "Wrong node pool size")
		// The cluster comes from the infra step: a pull request deploys this module in a step of its own.
		clusterId := tf.GetStep(t, e, "infra").String(t, "oke__cluster_id")
		assert.Equal(t, clusterId, *nodePool.ClusterId, "Node pool is not in the infra step's cluster")
	})
}

func getNodePool(t *testing.T, e *env.Environment, nodePoolId string) containerengine.NodePool {
	client, err := containerengine.NewContainerEngineClientWithConfigurationProvider(oracle.ConfigProvider())
	require.NoError(t, err, "Failed to create the Container Engine client")
	client.SetRegion(e.MustRegion(t))
	response, err := client.GetNodePool(context.Background(), containerengine.GetNodePoolRequest{NodePoolId: &nodePoolId})
	require.NoError(t, err, "Failed to get node pool %s", nodePoolId)
	return response.NodePool
}
