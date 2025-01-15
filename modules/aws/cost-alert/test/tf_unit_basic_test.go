package test

import (
	"testing"
	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/stretchr/testify/assert"
)

func TestCostAlert(t *testing.T) {
	t.Run("Biz", testTerraformCostAlertBiz)
	t.Run("Pri", testTerraformCostAlertPri)
}

func testTerraformCostAlertBiz(t *testing.T) {
	t.Parallel()
	outputs := aws.GetTFOutputs(t, "biz", "net")
	sns_topic_arns := aws.GetStringListValue(t, outputs, "cost-alert__sns_topic_arns")
	assert.NotEmpty(t, sns_topic_arns[0], "sns_topic_arns must not be empty")
}

func testTerraformCostAlertPri(t *testing.T) {
	t.Parallel()
	outputs := aws.GetTFOutputs(t, "pri", "net")
	assert.False(t, aws.HasKeyWithPrefix(t, outputs, "cost-alert__"), "Must not contain any cost-alert__ outputs.")
}
