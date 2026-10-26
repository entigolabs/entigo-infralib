package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/entigolabs/entigo-infralib-test/aws"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestEksNodeGroup(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)
		nodeGroupId := outputs.String(t, "eks-node-group__node_group_id")
		assert.NotEmpty(t, nodeGroupId, "Should not be empty: node_group_id")
		nodeGroupStatus := outputs.String(t, "eks-node-group__node_group_status")
		assert.Equal(t, nodeGroupStatus, "ACTIVE", "Should not be ACTIVE: node_group_status")
	})
}
