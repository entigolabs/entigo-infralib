package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

func TestExternalSecrets(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, c.Namespace, 10, 6*time.Second), "external-secrets deployment error")

		// The chart disables the webhook and cert-controller deployments
		// (webhook.create and certController.create are false), so they are
		// not waited for.

		_, err := k8s.WaitUntilClusterSecretStoreAvailable(t, c, c.Namespace, 10, 6*time.Second)
		require.NoError(t, err, "ClusterSecretStore not available error")
	})
}
