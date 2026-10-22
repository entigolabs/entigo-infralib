package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/entigolabs/entigo-infralib-test/aws"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

// The billing alarm needs us-east-1, which no environment provides yet, so
// this test skips until the module is part of one.
func TestCostAlert(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)
		snsTopicArns := outputs.StringList(t, "cost-alert__sns_topic_arns")
		assert.NotEmpty(t, snsTopicArns[0], "sns_topic_arns must not be empty")
	})
}
