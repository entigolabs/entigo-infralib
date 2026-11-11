package test

import (
	"testing"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

func TestRbacBindings(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		_ = k8s.Connect(t, e)
	})
}
