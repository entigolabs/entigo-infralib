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

func TestHelloWorld(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)
		helloWorld := outputs.String(t, "hello-world__hello_world")
		assert.Equal(t, helloWorld, fmt.Sprintf("Hello, %s-%s-hello-world!", e.Prefix, strings.ToLower(env.ModulePlacement(t, e).Step.Name)))
	})
}
