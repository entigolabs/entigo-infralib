package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/entigolabs/entigo-infralib-test/aws"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestCrossplane(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)
		iamRole := outputs.String(t, "crossplane__iam_role")
		assert.NotEmpty(t, iamRole, "iam_role must not be empty")
	})
}
