package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestTerraformGkeNodePool(t *testing.T) {
	t.Run("Pri", testTerraformGkeNodePoolPri)
	t.Run("Biz", testTerraformGkeNodePoolBiz)
}

func testTerraformGkeNodePoolPri(t *testing.T) {
	t.Parallel()
	testTerraformGkeNodePool(t, "pri")
}

func testTerraformGkeNodePoolBiz(t *testing.T) {
	t.Parallel()
	testTerraformGkeNodePool(t, "biz")
}

func testTerraformGkeNodePool(t *testing.T, envName string) {
	outputs := google.GetTFOutputs(t, envName)

	projectID := os.Getenv("GOOGLE_PROJECT")

	prefix := tf.GetStringValue(t, outputs, "gke-node-pool__prefix")
	assert.NotEmpty(t, prefix)

	nodePoolName := tf.GetStringValue(t, outputs, "gke-node-pool__node_pool_name")
	assert.NotEmpty(t, nodePoolName)
	assert.Contains(t, prefix, nodePoolName)

	clusterName := tf.GetStringValue(t, outputs, "gke-node-pool__cluster_name")
	assert.NotEmpty(t, clusterName)

	clusterRegion := tf.GetStringValue(t, outputs, "gke-node-pool__cluster_region")
	assert.NotEmpty(t, clusterRegion)

	nodePoolID := tf.GetStringValue(t, outputs, "gke-node-pool__node_pool_id")
	assert.NotEmpty(t, nodePoolID)
	assert.Contains(t, nodePoolID, fmt.Sprintf("projects/%s/locations/%s/clusters/%s/nodePools/%s", projectID, clusterRegion, clusterName, nodePoolName))
}
