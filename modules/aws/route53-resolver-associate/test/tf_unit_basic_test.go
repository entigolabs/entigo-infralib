package test

import (
	"testing"

	"github.com/entigolabs/entigo-infralib-test/env"
)

// The module has no outputs to check; a successful apply is the test.
func TestRoute53ResolverAssociate(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {})
}
