package test

import (
	"testing"

	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestTerraformConfigRules(t *testing.T) {
	t.Run("Biz", testTerraformConfigRulesBiz)
	t.Run("Pri", testTerraformConfigRulesPri)
}

func testTerraformConfigRulesBiz(t *testing.T) {
	t.Parallel()
	outputs := aws.GetTFOutputs(t, "biz")
	iam_role := tf.GetStringValue(t, outputs, "config-rules__iam_role")
	assert.NotEmpty(t, iam_role, "iam_role must not be empty")
}

func testTerraformConfigRulesPri(t *testing.T) {
	t.Parallel()
	outputs := aws.GetTFOutputs(t, "pri")
	iam_role := tf.GetStringValue(t, outputs, "config-rules__iam_role")
	assert.NotEmpty(t, iam_role, "iam_role must not be empty")
}
