package test

import (
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestK8sCrossplaneOracleDev(t *testing.T) {
	testK8sCrossplaneOracle(t, "oracle", "dev")
}

func testK8sCrossplaneOracle(t *testing.T, cloudName string, envName string) {
	t.Parallel()

	releaseName := "crossplane-oracle"

	kubectlOptions, _ := k8s.CheckKubectlConnection(t, cloudName, envName)
	// Crossplane runs provider deployments in its own install namespace, not in this
	// app's destination namespace (only the credentials secret lives there).
	kubectlOptions.Namespace = "crossplane-system"

	_, err := k8s.WaitUntilDeploymentRuntimeConfigAvailable(t, kubectlOptions, releaseName, 60, 2*time.Second)
	require.NoError(t, err, "DeploymentRuntimeConfigAvailable error")

	// The family provider must be installed and healthy before per-service providers.
	provider, err := k8s.WaitUntilProviderAvailable(t, kubectlOptions, "oracle-provider-family-oci", 120, 2*time.Second)
	require.NoError(t, err, "Provider family oci error")
	assert.NotNil(t, provider, "Provider family oci is nil")
	providerDeployment := k8s.GetStringValue(provider.Object, "status", "currentRevision")
	assert.NotEmpty(t, providerDeployment, "Provider family oci currentRevision is empty")
	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, providerDeployment, 60, 2*time.Second)

	// identity carries the per-app Policy MRs the other oracle modules depend on.
	identityProvider, err := k8s.WaitUntilProviderAvailable(t, kubectlOptions, "oracle-provider-oci-identity", 120, 2*time.Second)
	require.NoError(t, err, "Provider oci identity error")
	assert.NotNil(t, identityProvider, "Provider oci identity is nil")

	err = k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "oci.upbound.io/v1beta1", []string{"providerconfigs"}, 60, 2*time.Second)
	require.NoError(t, err, "Providerconfigs crd error")
	resource := schema.GroupVersionResource{Group: "oci.upbound.io", Version: "v1beta1", Resource: "providerconfigs"}
	_, err = k8s.WaitUntilProviderConfigAvailable(t, kubectlOptions, resource, releaseName, 60, 2*time.Second)
	require.NoError(t, err, "Provider config error")
}
