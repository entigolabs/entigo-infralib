package test

import (
	"testing"

	"github.com/entigolabs/entigo-infralib-common/oracle"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestTerraformOkeNodePool(t *testing.T) {
	t.Run("Biz", testTerraformOkeNodePoolBiz)
}

func testTerraformOkeNodePoolBiz(t *testing.T) {
	t.Parallel()
	outputs := oracle.GetTFOutputs(t, "biz")

	nodePoolId := tf.GetStringValue(t, outputs, "oke-node-pool__node_pool_id")
	assert.NotEmpty(t, nodePoolId, "node_pool_id was not returned")

	nodePoolName := tf.GetStringValue(t, outputs, "oke-node-pool__node_pool_name")
	assert.NotEmpty(t, nodePoolName, "node_pool_name was not returned")
}
