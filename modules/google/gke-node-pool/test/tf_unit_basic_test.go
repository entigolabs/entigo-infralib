package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/entigolabs/entigo-infralib-test/env"
	_ "github.com/entigolabs/entigo-infralib-test/google"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestGkeNodePool(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)

		projectID := e.MustProject(t)

		prefix := outputs.String(t, "gke-node-pool__prefix")
		assert.NotEmpty(t, prefix)

		nodePoolName := outputs.String(t, "gke-node-pool__node_pool_name")
		assert.NotEmpty(t, nodePoolName)
		assert.Contains(t, prefix, nodePoolName)

		clusterName := outputs.String(t, "gke-node-pool__cluster_name")
		assert.NotEmpty(t, clusterName)

		clusterRegion := outputs.String(t, "gke-node-pool__cluster_region")
		assert.NotEmpty(t, clusterRegion)

		nodePoolID := outputs.String(t, "gke-node-pool__node_pool_id")
		assert.NotEmpty(t, nodePoolID)
		assert.Contains(t, nodePoolID, fmt.Sprintf("projects/%s/locations/%s/clusters/%s/nodePools/%s", projectID, clusterRegion, clusterName, nodePoolName))
	})
}
