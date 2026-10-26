package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/oracle/oci-go-sdk/v65/keymanagement"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/oracle"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestKms(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)

		// oracle_biz.yaml keeps create_vault false, so the keys go into the vault the agent passes in,
		// which the agent names <prefix>-<region>.
		vaultId := outputs.String(t, "kms__vault_id")
		require.NotEmpty(t, vaultId, "vault_id was not returned")
		assert.Equal(t, fmt.Sprintf("%s-%s", e.Prefix, e.MustRegion(t)), outputs.String(t, "kms__vault_name"), "The keys are not in the agent's vault")

		managementEndpoint := outputs.String(t, "kms__vault_management_endpoint")
		require.NotEmpty(t, managementEndpoint, "vault_management_endpoint was not returned")
		assert.NotEmpty(t, outputs.String(t, "kms__vault_crypto_endpoint"), "vault_crypto_endpoint was not returned")

		// The three keys every cloud's kms module provides, created with the module's default shape.
		client := kmsManagementClient(t, managementEndpoint)
		keyPrefix := fmt.Sprintf("%s-%s-kms", e.Prefix, env.ModulePlacement(t, e).Step.Name)
		for _, name := range []string{"data", "config", "telemetry"} {
			keyId := outputs.String(t, "kms__"+name+"_key_id")
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

		// oracle_biz.yaml turns the CA key off - see the comment there for why it is not exercised here.
		assert.Empty(t, outputs.String(t, "kms__ca_key_id"), "ca_key_id should be empty when create_ca_key is false")
	})
}

// The management client is bound to the vault's own endpoint rather than a region.
func kmsManagementClient(t *testing.T, managementEndpoint string) keymanagement.KmsManagementClient {
	client, err := keymanagement.NewKmsManagementClientWithConfigurationProvider(oracle.ConfigProvider(), managementEndpoint)
	require.NoError(t, err, "Failed to create the KMS management client")
	return client
}
