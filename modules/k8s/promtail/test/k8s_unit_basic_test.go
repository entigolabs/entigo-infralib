package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

// The module is part of no environment, so this skips until it is added to one.
func TestPromtail(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		dsname, err := k8s.GetDaemonSetE(t, c, c.Namespace)
		require.NoError(t, err, "daemonset %s error", c.Namespace)
		assert.NotEmpty(t, dsname, "Daemonset was not returned")
	})
}
