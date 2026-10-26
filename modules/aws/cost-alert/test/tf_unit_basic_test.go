package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/entigolabs/entigo-infralib-test/aws"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

// The billing alarm reads the AWS/Billing metric, which exists in us-east-1
// only; the aws_us environment lives there (environments/aws_us.env).
func TestCostAlert(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)
		snsTopicArns := outputs.StringList(t, "cost-alert__sns_topic_arns")
		assert.NotEmpty(t, snsTopicArns[0], "sns_topic_arns must not be empty")
	})
}
