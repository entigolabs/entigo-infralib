package test

import (
	"testing"

	"github.com/entigolabs/entigo-infralib-common/azure"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestTerraformAksNodePool(t *testing.T) {
	t.Run("Biz", testTerraformAksNodePoolBiz)
	t.Run("Pri", testTerraformAksNodePoolPri)
}

func checkAksNodePool(t *testing.T, env string) {
	outputs := azure.GetTFOutputs(t, env)
	infraOutputs := azure.GetTFOutputsStep(t, env, "infra")

	assert.Equal(t, "aksnodepool", tf.GetStringValue(t, outputs, "aks-node-pool__node_pool_name"))
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "aks-node-pool__node_pool_id"), "node_pool_id was not returned")
	assert.Equal(t, tf.GetStringValue(t, infraOutputs, "aks__cluster_id"), tf.GetStringValue(t, outputs, "aks-node-pool__cluster_id"), "cluster_id must be the infra aks cluster")
}

func testTerraformAksNodePoolBiz(t *testing.T) {
	t.Parallel()
	checkAksNodePool(t, "biz")
}

func testTerraformAksNodePoolPri(t *testing.T) {
	t.Parallel()
	checkAksNodePool(t, "pri")
}
