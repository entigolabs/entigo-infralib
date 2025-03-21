package test

import (
	"testing"
	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestCostAlert(t *testing.T) {
	t.Run("Us", testTerraformCostAlertUs)
}

func testTerraformCostAlertUs(t *testing.T) {
	t.Parallel()
	outputs := aws.GetTFOutputs(t, "us")
	sns_topic_arns := tf.GetStringListValue(t, outputs, "cost-alert__sns_topic_arns")
	assert.NotEmpty(t, sns_topic_arns[0], "sns_topic_arns must not be empty")
}
