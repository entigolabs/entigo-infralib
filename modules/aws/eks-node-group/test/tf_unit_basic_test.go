package test

import (
	"testing"
	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestTerraformEks(t *testing.T) {
	t.Run("Biz", testTerraformEksBiz)
	t.Run("Pri", testTerraformEksPri)
}

func testTerraformEksBiz(t *testing.T) {
	t.Parallel()
	outputs := aws.GetTFOutputs(t, "biz")
	node_group_id := tf.GetStringValue(t, outputs, "eks-node-group__node_group_id")
	assert.NotEmpty(t, node_group_id, "Should not be empty: node_group_id")
	node_group_status := tf.GetStringValue(t, outputs, "eks-node-group__node_group_status")
	assert.Equal(t, node_group_status, "ACTIVE", "Should not be ACTIVE: node_group_status")

}

func testTerraformEksPri(t *testing.T) {
	t.Parallel()
	outputs := aws.GetTFOutputs(t, "pri")
	node_group_id := tf.GetStringValue(t, outputs, "eks-node-group__node_group_id")
	assert.NotEmpty(t, node_group_id, "Should not be empty: node_group_id")
	node_group_status := tf.GetStringValue(t, outputs, "eks-node-group__node_group_status")
	assert.Equal(t, node_group_status, "ACTIVE", "Should not be ACTIVE: node_group_status")

}
