package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

func TestEntigoPortalAgent(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		err := k8s.WaitUntilDeploymentAvailableE(t, c, c.Namespace, 30, 10*time.Second)
		require.NoError(t, err, "%s deployment error", c.Namespace)
	})
}
