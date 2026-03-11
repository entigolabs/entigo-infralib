package test

import (
	"testing"

	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestTerraformEfs(t *testing.T) {
	t.Run("Biz", testTerraformEfsBiz)
	t.Run("Pri", testTerraformEfsPri)
}

func testTerraformEfsBiz(t *testing.T) {
	t.Parallel()
	outputs := aws.GetTFOutputs(t, "biz")
	// Biz should have 2 EFS instances
	efsIds := tf.GetStringListValue(t, outputs, "efs__id")
	assert.Equal(t, 2, len(efsIds), "Expected 2 EFS IDs in biz")
}

func testTerraformEfsPri(t *testing.T) {
	t.Parallel()
	outputs := aws.GetTFOutputs(t, "pri")
	// Pri should have 1 EFS instance
	efsIds := tf.GetStringListValue(t, outputs, "efs__id")
	assert.Equal(t, 1, len(efsIds), "Expected 1 EFS ID in pri")
}
