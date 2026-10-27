package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/entigolabs/entigo-infralib-test/aws"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestEfs(t *testing.T) {
	env.RunEach(t, map[string]env.TestFunc{
		"aws_biz": testEfsBiz,
		"aws_pri": testEfsPri,
	})
}

func testEfsBiz(t *testing.T, e *env.Environment) {
	outputs := tf.Get(t, e)
	// Biz should have 2 EFS instances
	efsIds := outputs.StringList(t, "efs__id")
	assert.Equal(t, 2, len(efsIds), "Expected 2 EFS IDs in biz")
}

func testEfsPri(t *testing.T, e *env.Environment) {
	outputs := tf.Get(t, e)
	// Pri should have 1 EFS instance
	efsIds := outputs.StringList(t, "efs__id")
	assert.Equal(t, 1, len(efsIds), "Expected 1 EFS ID in pri")
}
