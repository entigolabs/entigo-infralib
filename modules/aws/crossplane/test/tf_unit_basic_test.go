package test

import (
	"testing"
	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestTerraformCrossplane(t *testing.T) {
	t.Run("Biz", testTerraformCrossplaneBiz)
	t.Run("Pri", testTerraformCrossplanePri)
}

func testTerraformCrossplaneBiz(t *testing.T) {
	t.Parallel()
	outputs := aws.GetTFOutputs(t, "biz", "infra")
	iam_role := tf.GetStringValue(t, outputs, "crossplane__iam_role")
	assert.NotEmpty(t, iam_role, "iam_role must not be empty")
}

func testTerraformCrossplanePri(t *testing.T) {
	t.Parallel()
	outputs := aws.GetTFOutputs(t, "pri", "infra")
	iam_role := tf.GetStringValue(t, outputs, "crossplane__iam_role")
	assert.NotEmpty(t, iam_role, "iam_role must not be empty")
}
