package test

import (
	"testing"

	"github.com/entigolabs/entigo-infralib-common/oracle"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestTerraformOke(t *testing.T) {
	t.Run("Biz", testTerraformOkeBiz)
}

func testTerraformOkeBiz(t *testing.T) {
	t.Parallel()
	outputs := oracle.GetTFOutputs(t, "biz")

	clusterId := tf.GetStringValue(t, outputs, "oke__cluster_id")
	assert.NotEmpty(t, clusterId, "cluster_id was not returned")

	clusterName := tf.GetStringValue(t, outputs, "oke__cluster_name")
	assert.NotEmpty(t, clusterName, "cluster_name was not returned")

	kubernetesVersion := tf.GetStringValue(t, outputs, "oke__kubernetes_version")
	assert.NotEmpty(t, kubernetesVersion, "kubernetes_version was not returned")

	publicEndpoint := tf.GetStringValue(t, outputs, "oke__public_endpoint")
	assert.NotEmpty(t, publicEndpoint, "public_endpoint was not returned")

	kubernetesEndpoint := tf.GetStringValue(t, outputs, "oke__kubernetes_endpoint")
	assert.NotEmpty(t, kubernetesEndpoint, "kubernetes_endpoint was not returned")

	mainNodePoolId := tf.GetStringValue(t, outputs, "oke__main_node_pool_id")
	assert.NotEmpty(t, mainNodePoolId, "main_node_pool_id was not returned")

	monNodePoolId := tf.GetStringValue(t, outputs, "oke__mon_node_pool_id")
	assert.NotEmpty(t, monNodePoolId, "mon_node_pool_id was not returned")

	toolsNodePoolId := tf.GetStringValue(t, outputs, "oke__tools_node_pool_id")
	assert.NotEmpty(t, toolsNodePoolId, "tools_node_pool_id was not returned")
}
