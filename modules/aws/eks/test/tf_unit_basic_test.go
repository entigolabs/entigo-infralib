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

func TestEks(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)
		clusterName := outputs.String(t, "eks__cluster_name")
		assert.Equal(t, fmt.Sprintf("%s-%s-eks", e.Prefix, strings.ToLower(env.ModulePlacement(t, e).Step.Name)), clusterName, "Wrong cluster_name returned")
	})
}
