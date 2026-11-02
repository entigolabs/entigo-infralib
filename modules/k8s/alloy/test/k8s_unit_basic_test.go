package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

func TestAlloy(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		logsDaemonSetName := fmt.Sprintf("%s-logs", c.Namespace)

		daemonSet, err := k8s.GetDaemonSetE(t, c, logsDaemonSetName)
		require.NoError(t, err, "Daemonset %s error", logsDaemonSetName)
		assert.NotEmpty(t, daemonSet, "Daemonset was not returned")
	})
}
