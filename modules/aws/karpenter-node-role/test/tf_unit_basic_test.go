package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/entigolabs/entigo-infralib-test/aws"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestKarpenterNodeRole(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)
		roleName := outputs.String(t, "karpenter-node-role__role_name")
		assert.Equal(t, fmt.Sprintf("%s-%s-karpenter-node-role", e.Prefix, strings.ToLower(env.ModulePlacement(t, e).Step.Name)), roleName, "Wrong role_name returned")
		roleArn := outputs.String(t, "karpenter-node-role__role_arn")
		assert.NotEmpty(t, roleArn, "Empty role_arn returned")
	})
}
