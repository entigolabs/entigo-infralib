package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

// Crossplane runs provider deployments in its own install namespace (the
// crossplane-core application), not in this app's namespace.
const crossplaneNamespace = "crossplane-system"

func TestCrossplaneKafka(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)
		releaseName := "contrib-provider-kafka"

		_, err := k8s.WaitUntilDeploymentRuntimeConfigAvailable(t, c, releaseName, 60, 1*time.Second)
		require.NoError(t, err, "DeploymentRuntimeConfigAvailable error")

		// Install AWS provider
		provider, err := k8s.WaitUntilProviderAvailable(t, c, releaseName, 60, 1*time.Second)
		require.NoError(t, err, "Provider aws error")
		assert.NotNil(t, provider, "Provider aws is nil")
		providerDeployment := k8s.GetStringValue(provider.Object, "status", "currentRevision")
		assert.NotEmpty(t, providerDeployment, "Provider aws currentRevision is empty")
		k8s.WaitUntilDeploymentAvailable(t, c.WithNamespace(crossplaneNamespace), providerDeployment, 60, 1*time.Second)
	})
}
