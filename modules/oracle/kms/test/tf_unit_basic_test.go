package test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/oracle"
	"github.com/entigolabs/entigo-infralib-common/tf"
	ocicommon "github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/keymanagement"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerraformKms(t *testing.T) {
	t.Run("Biz", testTerraformKmsBiz)
}

func testTerraformKmsBiz(t *testing.T) {
	t.Parallel()
	prefix := "biz"
	outputs := oracle.GetTFOutputs(t, prefix)

	// biz.yaml keeps create_vault false, so the keys go into the vault the agent passes in,
	// which the agent names <prefix>-<region>.
	vaultId := tf.GetStringValue(t, outputs, "kms__vault_id")
	require.NotEmpty(t, vaultId, "vault_id was not returned")
	assert.Equal(t, fmt.Sprintf("%s-%s", prefix, os.Getenv("OCI_REGION")), tf.GetStringValue(t, outputs, "kms__vault_name"), "The keys are not in the agent's vault")

	managementEndpoint := tf.GetStringValue(t, outputs, "kms__vault_management_endpoint")
	require.NotEmpty(t, managementEndpoint, "vault_management_endpoint was not returned")
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "kms__vault_crypto_endpoint"), "vault_crypto_endpoint was not returned")

	// The three keys every cloud's kms module provides, created with the module's default shape.
	client := kmsManagementClient(t, managementEndpoint)
	keyPrefix := fmt.Sprintf("%s-%s-kms", prefix, strings.ToLower(os.Getenv("STEP_NAME")))
	for _, name := range []string{"data", "config", "telemetry"} {
		keyId := tf.GetStringValue(t, outputs, "kms__"+name+"_key_id")
		require.True(t, strings.HasPrefix(keyId, "ocid1.key."), "Wrong value for %s_key_id returned", name)

		response, err := client.GetKey(context.Background(), keymanagement.GetKeyRequest{KeyId: &keyId})
		require.NoError(t, err, "Failed to get the %s key", name)
		key := response.Key
		assert.True(t, strings.HasPrefix(*key.DisplayName, keyPrefix+"-"+name+"-"), "Wrong name for the %s key: %s", name, *key.DisplayName)
		assert.Equal(t, vaultId, *key.VaultId, "The %s key is not in the vault", name)
		assert.Equal(t, keymanagement.KeyLifecycleStateEnabled, key.LifecycleState, "The %s key is not enabled", name)
		assert.Equal(t, keymanagement.KeyShapeAlgorithmAes, key.KeyShape.Algorithm, "Wrong algorithm for the %s key", name)
		assert.Equal(t, 32, *key.KeyShape.Length, "Wrong length for the %s key", name)
		assert.Equal(t, keymanagement.KeyProtectionModeSoftware, key.ProtectionMode, "Wrong protection mode for the %s key", name)
	}

	// biz.yaml turns the CA key off - see the comment there for why it is not exercised here.
	assert.Empty(t, tf.GetStringValue(t, outputs, "kms__ca_key_id"), "ca_key_id should be empty when create_ca_key is false")
}

func kmsManagementClient(t *testing.T, managementEndpoint string) keymanagement.KmsManagementClient {
	provider := ocicommon.DefaultConfigProvider()
	if configFile := os.Getenv("OCI_CONFIG_FILE"); configFile != "" {
		provider = ocicommon.CustomProfileConfigProvider(configFile, "DEFAULT")
	}
	client, err := keymanagement.NewKmsManagementClientWithConfigurationProvider(provider, managementEndpoint)
	require.NoError(t, err, "Failed to create the KMS management client")
	return client
}
